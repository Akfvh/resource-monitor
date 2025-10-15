package pseudo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"context"
	"time"

	"golang.org/x/sys/unix"

	T "resmon/pkg/types"
)

// PSI 스코프/경로를 유연하게
type PSIScope struct {
	Scope    string // "system"|"cgroup"
	CgPath   string // cgroup 압박 파일들이 있는 디렉터리
}

// 기본 스코프 자동 감지(실패 시 system)
func DefaultPSIScope() PSIScope {
	// 간단 버전: system으로
	return PSIScope{Scope: "system", CgPath: "/sys/fs/cgroup"}
}

func psiFilePath(scope PSIScope, res string) string {
	if scope.Scope == "cgroup" {
		return filepath.Join(scope.CgPath, res+".pressure")
	}
	return "/proc/pressure/" + res
}

func parsePSILine(line string) (avg10, avg60, avg300 float64, totalUs uint64) {
	parts := strings.Fields(line)
	m := map[string]string{}
	for _, p := range parts[1:] {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	f := func(k string) float64 {
		v, _ := strconv.ParseFloat(strings.TrimSpace(m[k]), 64)
		return v
	}
	u := func(k string) uint64 {
		v, _ := strconv.ParseUint(strings.TrimSpace(m[k]), 10, 64)
		return v
	}
	return f("avg10"), f("avg60"), f("avg300"), u("total")
}

func readPSIFile(path string, res string) (some T.PSIEvent, full T.PSIEvent, err error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return some, full, e
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	now := T.NowMS()
	for _, ln := range lines {
		if strings.HasPrefix(ln, "some ") {
			a10, a60, a300, tot := parsePSILine(ln)
			some = T.PSIEvent{Res: res, Kind: "some", Avg10: a10, Avg60: a60, Avg300: a300, TotalUs: tot, Ts: now}
		} else if strings.HasPrefix(ln, "full ") {
			a10, a60, a300, tot := parsePSILine(ln)
			full = T.PSIEvent{Res: res, Kind: "full", Avg10: a10, Avg60: a60, Avg300: a300, TotalUs: tot, Ts: now}
		}
	}
	return
}

// 커널 PSI 트리거 + poll 기반 이벤트 채널
func SpawnPSIWatcher(ctx context.Context, scope PSIScope, res, kind string, thrUs, winUs int) (<-chan T.PSIEvent, error) {
	out := make(chan T.PSIEvent, 16)

	path := psiFilePath(scope, res)
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	// 트리거 쓰기
	trig := fmt.Sprintf("%s %d %d\n", kind, thrUs, winUs)
	if _, err := unix.Write(fd, []byte(trig)); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}

	go func() {
		defer unix.Close(fd)
		defer close(out)
		pfd := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLPRI}}

		for {
			// ctx 취소 처리
			select {
			case <-ctx.Done():
				return
			default:
			}

			_, err := unix.Poll(pfd, -1) // 1s 타임아웃(취소 체크용)
			if err != nil { return }
			re := pfd[0].Revents
			if re&(unix.POLLERR|unix.POLLNVAL) != 0 {
				return
			}
			if re&unix.POLLPRI != 0 {
				s, f, _ := readPSIFile(path, res)
				ev := s
				if kind == "full" {
					ev = f
				}
				ev.Kind = kind
				ev.Threshold = thrUs
				ev.Window = winUs
				select {
				case out <- ev:
				default: // 채널 가득이면 드랍(최신값 우선)
				}
			}
		}
	}()
	return out, nil
}

// 일정 주기로 /proc/pressure/*를 읽어 최신 Avg10을 보장하는 간단 폴러
func SpawnPSIPoller(ctx context.Context, scope PSIScope, res string, every time.Duration) <-chan T.PSIEvent {
	out := make(chan T.PSIEvent, 1)
	path := psiFilePath(scope, res)
	go func() {
		defer close(out)
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s, _, err := readPSIFile(path, res)
				if err == nil {
					select {
					case out <- s:
					default:
					}
				}
			}
		}
	}()
	return out
}
