# Resource Monitor

A Go-based resource monitor with linux pseudo-files and perf tool

## Functionalities

- Network Bandwidth
- LLC MPKI and Memory Bandwidth
- CPU/IO/MEMORY PSI (Pressure Stall Information)

## Installations

```bash
# Download Dependencies
go mod tidy

# Build
go build -o resmon ./cmd/resmon
```

## Usage

```bash
# Default Usage
sudo ./resmon

# Specify Configuration file
sudo ./resmon -config /path/to/config.yaml

# help
sudo ./resmon -h
```

### Config.yaml

```yaml
# Resource Monitor Configuration
monitoring:
  # Network Monitor
  network:
    interface: "enp4s0"
    interval: "1s"
  
  # PSI Monitor
  psi:
    memory:
      threshold_us: 150000
      window_us: 1000000
      kind: "some"
    cpu:
      threshold_us: 100000
      window_us: 1000000
      kind: "some"
    io:
      threshold_us: 150000
      window_us: 1000000
      kind: "full"
    memory_poll_interval: "1s"
  
  # Perf Monitoring
  perf:
    interval: "1s"
    events:
      - "LLC-loads"
      - "LLC-load-misses"
      - "LLC-stores"
      - "LLC-store-misses"
      - "instructions"
      - "unc_m_cas_count_rd"
      - "unc_m_cas_count_wr"

# Output
output:
  console: true
  log_level: "info"
  metrics_interval: "1s"

# PSI scope
psi_scope:
  type: "system"  # "system" or "cgroup"
  cgroup_path: "/sys/fs/cgroup"
```

## Configurations

### Network Monitor
- `interface`: Network Interface to Monitor (Default: "enp4s0")
- `interval`: Sampling Interval (Ex: "1s", "500ms")

### PSI Monitor
- `threshold_us`: PSI threshold (microsecond)
- `window_us`: PSI window (microsecond)
- `kind`: pressure kind ("some" | "full")
- `memory_poll_interval`: Memory polling interval

### Perf Monitor
- `interval`: perf sampling interval
- `events`: perf events to monitor

## Sample Output

```
[PSI] mem some avg10=2.45%
[PSI] cpu some avg10=1.23%
[PSI] io full avg10=0.87%
[NET] enp4s0 rx=1024000B/s tx=512000B/s
[PERF] MemBW total=1250.5MB/s (R=800.2 W=450.3)
[PERF] LLC mpki=15.67 hit=0.85 loads=125000 stores=75000
```

## Requirements

- Linux kernel 4.20+ (PSI support)
- perf tool
  - perf support for specified events

## 라이선스

MIT License
