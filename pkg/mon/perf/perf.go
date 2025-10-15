package perf

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	T "resmon/pkg/types"
)

type Config struct {
	Interval time.Duration
	Events   []string // perf 이벤트 이름들
}

// 기본 이벤트(LLC+메모리 BW)
func DefaultConfig(interval time.Duration) Config {
	return Config{
		Interval: interval,
		Events: []string{
			"LLC-loads",
			"LLC-load-misses",
			"LLC-stores",
			"LLC-store-misses",
			"instructions",
			"unc_m_cas_count_rd",
			"unc_m_cas_count_wr",
		},
	}
}

// perf 한 프로세스로 LLC + MemBW 동시 파싱
// 반환: membw 채널, llc 채널
func SpawnPerfMonitor(ctx context.Context, cfg Config) (<-chan T.MemBw, <-chan T.LLCSample, error) {
	memCh := make(chan T.MemBw, 8)
	llcCh := make(chan T.LLCSample, 8)

	args := []string{
		"stat", "-a",
		"-I", strconv.Itoa(int(cfg.Interval / time.Millisecond)),
		"-x", ",",
		"-e", strings.Join(cfg.Events, ","),
		"--", "sleep", "1000000",
	}

	cmd := exec.CommandContext(ctx, "perf", args...)
	stdout, err := cmd.StderrPipe() // perf는 stderr에 출력
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	go func() {
		defer close(memCh)
		defer close(llcCh)
		defer func() { _ = cmd.Process.Kill() }()

		sc := bufio.NewScanner(stdout)

		// 현재 틱 누적 변수
		var loads, lmiss, stores, smiss, instr uint64
		var haveL, haveLM, haveS, haveSM, haveI bool
		var rd, wr float64
		var haveRD, haveWR bool
		sec := float64(cfg.Interval) / float64(time.Second)

		reset := func() {
			loads, lmiss, stores, smiss, instr = 0, 0, 0, 0, 0
			haveL, haveLM, haveS, haveSM, haveI = false, false, false, false, false
			rd, wr = 0, 0
			haveRD, haveWR = false, false
		}

		for sc.Scan() {
			cols := strings.Split(sc.Text(), ",")
			// perf -x, 포맷: time, value, unit, event, runtime, CPUs
			// 최소 4컬럼 방어
			if len(cols) < 4 {
				continue
			}
			valStr := strings.TrimSpace(cols[1])
			ev := strings.TrimSpace(cols[3])
			if valStr == "" || strings.Contains(valStr, "not counted") || strings.Contains(ev, "duration_time") {
				continue
			}
			// value는 샘플링 간격 동안의 증분
			if strings.Contains(ev, "LLC-loads") {
				if v, e := strconv.ParseUint(valStr, 10, 64); e == nil {
					loads += v; haveL = true
				}
			} else if strings.Contains(ev, "LLC-load-misses") {
				if v, e := strconv.ParseUint(valStr, 10, 64); e == nil {
					lmiss += v; haveLM = true
				}
			} else if strings.Contains(ev, "LLC-stores") {
				if v, e := strconv.ParseUint(valStr, 10, 64); e == nil {
					stores += v; haveS = true
				}
			} else if strings.Contains(ev, "LLC-store-misses") {
				if v, e := strconv.ParseUint(valStr, 10, 64); e == nil {
					smiss += v; haveSM = true
				}
			} else if ev == "instructions" {
				if v, e := strconv.ParseUint(valStr, 10, 64); e == nil {
					instr = v; haveI = true
				}
			} else if strings.Contains(ev, "cas_count_rd") || strings.Contains(ev, "cas_count_read") {
				if v, e := strconv.ParseFloat(valStr, 64); e == nil {
					rd = v; haveRD = true
				}
			} else if strings.Contains(ev, "cas_count_wr") || strings.Contains(ev, "cas_count_write") {
				if v, e := strconv.ParseFloat(valStr, 64); e == nil {
					wr = v; haveWR = true
				}
			}

			// 한 틱 완료 조건: 최소 instructions를 만난 시점으로 가정
			if haveI {
				// LLC 샘플
				if haveL || haveLM || haveS || haveSM {
					totAcc := loads + stores
					totMiss := lmiss + smiss
					var mpki, hit float64
					if instr > 0 {
						mpki = 1000.0 * float64(totMiss) / float64(instr)
					}
					if totAcc > 0 {
						hit = 1.0 - float64(totMiss)/float64(totAcc)
					}
					llc := T.LLCSample{
						MPKI: mpki, HitRate: hit,
						Loads: loads, Stores: stores, Misses: totMiss,
						Instr: instr, Ts: T.NowMS(), Source: "perf",
					}
					select { case llcCh <- llc: default: }
				}

				// MemBW 샘플
				if haveRD && haveWR {
					totalMBs := ((rd + wr) * 64.0) / (1024.0 * 1024.0) / sec
					readMBs := (rd * 64.0) / (1024.0 * 1024.0) / sec
					writeMBs := (wr * 64.0) / (1024.0 * 1024.0) / sec
					mb := T.MemBw{
						Source: "perf", ReadMBs: readMBs, WriteMBs: writeMBs,
						TotalMBs: totalMBs, Ts: T.NowMS(),
					}
					select { case memCh <- mb: default: }
				}
				reset()
			}
		}
	}()
	return memCh, llcCh, nil
}
