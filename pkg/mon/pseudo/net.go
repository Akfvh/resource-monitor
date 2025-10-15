package pseudo

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"strings"
	"fmt"

	T "resmon/pkg/types"
)

func readUintFrom(path string) (uint64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := string(b)
	var v uint64
	_, err = fmtSscanf(s, "%d", &v)
	if err != nil {
		// fallback: ParseUint
		return strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	}
	return v, nil
}

// 작은 의존 제거용
func fmtSscanf(s, f string, a ...any) (int, error) { return fmt.Sscanf(s, f, a...) }

func SpawnNetWatcher(ctx context.Context, iface string, interval time.Duration) (<-chan T.NetSample, error) {
	base := "/sys/class/net/" + iface + "/statistics"
	_, err := os.Stat(base)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	out := make(chan T.NetSample, 8)
	go func() {
		defer close(out)
		rxPrev, _ := readUintFrom(filepath.Join(base, "rx_bytes"))
		txPrev, _ := readUintFrom(filepath.Join(base, "tx_bytes"))
		prevT := time.Now()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				rx, _ := readUintFrom(filepath.Join(base, "rx_bytes"))
				tx, _ := readUintFrom(filepath.Join(base, "tx_bytes"))
				now := time.Now()
				dt := now.Sub(prevT).Seconds()
				var rbps, tbps uint64
				if rx >= rxPrev && tx >= txPrev && dt > 0 {
					rbps = uint64(float64(rx-rxPrev) / dt)
					tbps = uint64(float64(tx-txPrev) / dt)
				}
				prevT, rxPrev, txPrev = now, rx, tx
				ns := T.NetSample{Iface: iface, RxBps: rbps, TxBps: tbps, Ts: T.NowMS()}
				select {
				case out <- ns:
				default:
				}
			}
		}
	}()
	return out, nil
}
