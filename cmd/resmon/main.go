package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"resmon/pkg/config"
	P "resmon/pkg/mon/pseudo"
	X "resmon/pkg/mon/perf"
	T "resmon/pkg/types"
)

func getenv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		fmt.Println("Using default configuration...")
		cfg = config.GetDefaultConfig()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Get intervals from config
	netInterval, err := cfg.GetNetworkInterval()
	if err != nil {
		fmt.Printf("Invalid network interval: %v, using 1s\n", err)
		netInterval = time.Second
	}

	psiMemPollInterval, err := cfg.GetPSIMemoryPollInterval()
	if err != nil {
		fmt.Printf("Invalid PSI memory poll interval: %v, using 1s\n", err)
		psiMemPollInterval = time.Second
	}

	perfInterval, err := cfg.GetPerfInterval()
	if err != nil {
		fmt.Printf("Invalid perf interval: %v, using 1s\n", err)
		perfInterval = time.Second
	}

	// 1) pseudo-file 모듈 (PSI + NIC)
	scope := P.PSIScope{
		Scope:    cfg.PSIScope.Type,
		CgPath:   cfg.PSIScope.CgroupPath,
	}
	
	psiMemEv, _ := P.SpawnPSIWatcher(ctx, scope, "memory", cfg.Monitoring.PSI.Memory.Kind,
		cfg.Monitoring.PSI.Memory.ThresholdUs, cfg.Monitoring.PSI.Memory.WindowUs)
	psiCpuEv, _ := P.SpawnPSIWatcher(ctx, scope, "cpu", cfg.Monitoring.PSI.CPU.Kind,
		cfg.Monitoring.PSI.CPU.ThresholdUs, cfg.Monitoring.PSI.CPU.WindowUs)
	psiIoEv, _ := P.SpawnPSIWatcher(ctx, scope, "io", cfg.Monitoring.PSI.IO.Kind,
		cfg.Monitoring.PSI.IO.ThresholdUs, cfg.Monitoring.PSI.IO.WindowUs)
	psiMemPoll := P.SpawnPSIPoller(ctx, scope, "memory", psiMemPollInterval)

	netCh, err := P.SpawnNetWatcher(ctx, cfg.Monitoring.Network.Interface, netInterval)
	if err != nil {
		fmt.Println("net watcher error:", err)
	}

	// 2) perf 모듈 (LLC + MemBW)
	perfConfig := X.Config{
		Interval: perfInterval,
		Events:   cfg.Monitoring.Perf.Events,
	}
	memCh, llcCh, err := X.SpawnPerfMonitor(ctx, perfConfig)
	if err != nil {
		fmt.Println("perf monitor error:", err)
	}

	// Get metrics interval from config
	metricsInterval, err := cfg.GetMetricsInterval()
	if err != nil {
		fmt.Printf("Invalid metrics interval: %v, using 1s\n", err)
		metricsInterval = time.Second
	}

	// 샘플 콘솔 출력 (실전에서는 UDP 송신/집계기로 연결)
	tick := time.NewTicker(metricsInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case e := <-psiMemEv:
			fmt.Printf("[PSI] mem some avg10=%.2f%%\n", e.Avg10)
		case e := <-psiCpuEv:
			fmt.Printf("[PSI] cpu some avg10=%.2f%%\n", e.Avg10)
		case e := <-psiIoEv:
			fmt.Printf("[PSI] io full avg10=%.2f%%\n", e.Avg10)
		case e := <-psiMemPoll:
			_ = e // 필요하면 사용 (정규화: e.Avg10/100)
		case n := <-netCh:
			fmt.Printf("[NET] %s rx=%dB/s tx=%dB/s\n", n.Iface, n.RxBps, n.TxBps)
		case m := <-memCh:
			fmt.Printf("[PERF] MemBW total=%.0fMB/s (R=%.0f W=%.0f)\n", m.TotalMBs, m.ReadMBs, m.WriteMBs)
		case l := <-llcCh:
			fmt.Printf("[PERF] LLC mpki=%.2f hit=%.2f loads=%d stores=%d\n", l.MPKI, l.HitRate, l.Loads, l.Stores)
		case <-tick.C:
			// 주기 스냅샷/스코어링 등을 여기서
			_ = T.NowMS()
		}
	}
}
