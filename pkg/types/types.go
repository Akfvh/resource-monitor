package types

import "time"

// 공용 타입들

type PSIEvent struct {
	Res       string  `json:"res"` // cpu|memory|io
	Kind      string  `json:"kind"` // some|full
	Threshold int     `json:"thr_us"`
	Window    int     `json:"win_us"`
	Ts        int64   `json:"ts_unix_ms"`
	Avg10     float64 `json:"avg10"`  // NOTE: /proc 값 그대로(%) → 사용하는 쪽에서 /100 정규화 권장
	Avg60     float64 `json:"avg60"`
	Avg300    float64 `json:"avg300"`
	TotalUs   uint64  `json:"total_us"`
}

type NetSample struct {
	Iface string `json:"iface"`
	RxBps uint64 `json:"rx_bps"`
	TxBps uint64 `json:"tx_bps"`
	Ts    int64  `json:"ts_unix_ms"`
}

type MemBw struct {
	Source   string  `json:"source"` // perf
	ReadMBs  float64 `json:"read_mbps"`
	WriteMBs float64 `json:"write_mbps"`
	TotalMBs float64 `json:"total_mbps"`
	Ts       int64   `json:"ts_unix_ms"`
}

type LLCSample struct {
	MPKI   float64 `json:"mpki"`
	HitRate float64 `json:"hit_rate"`
	Loads  uint64  `json:"loads"`
	Stores uint64  `json:"stores"`
	Misses uint64  `json:"misses"` // load+store misses
	Instr  uint64  `json:"instructions"`
	Ts     int64   `json:"ts_unix_ms"`
	Source string  `json:"source"` // perf
}

func NowMS() int64 { return time.Now().UnixMilli() }
