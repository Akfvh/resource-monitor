# Resource Monitor

네트워크 대역폭, 성능 LLC MPKI, 그리고 PSI 모니터링을 위한 Go 기반 리소스 모니터입니다.

## 기능

- **네트워크 모니터링**: `/sys/class/net` 기반 실시간 대역폭 측정
- **PSI 모니터링**: 커널 PSI 트리거 기반 압박 상태 감지
- **성능 모니터링**: perf 도구를 통한 LLC MPKI 및 메모리 대역폭 측정
- **설정 파일 지원**: YAML 기반 유연한 설정 관리

## 설치

```bash
# 의존성 다운로드
go mod tidy

# 빌드
go build -o resmon ./cmd/resmon
```

## 사용법

### 기본 사용법

```bash
# 기본 설정으로 실행
./resmon

# 설정 파일 지정
./resmon -config /path/to/config.yaml

# 도움말
./resmon -h
```

### 설정 파일

`config.yaml` 파일을 통해 모든 모니터링 설정을 관리할 수 있습니다:

```yaml
# Resource Monitor Configuration
monitoring:
  # 네트워크 모니터링 설정
  network:
    interface: "enp4s0"
    interval: "1s"
  
  # PSI 모니터링 설정
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
  
  # 성능 모니터링 설정
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

# 출력 설정
output:
  console: true
  log_level: "info"
  metrics_interval: "1s"

# PSI 스코프 설정
psi_scope:
  type: "system"  # "system" or "cgroup"
  cgroup_path: "/sys/fs/cgroup"
```

## 설정 옵션

### 네트워크 모니터링
- `interface`: 모니터링할 네트워크 인터페이스 (기본값: "enp4s0")
- `interval`: 샘플링 간격 (예: "1s", "500ms")

### PSI 모니터링
- `threshold_us`: 압박 임계값 (마이크로초)
- `window_us`: 측정 윈도우 (마이크로초)
- `kind`: 압박 유형 ("some" 또는 "full")
- `memory_poll_interval`: 메모리 폴링 간격

### 성능 모니터링
- `interval`: perf 샘플링 간격
- `events`: 측정할 perf 이벤트 목록

### 출력 설정
- `console`: 콘솔 출력 활성화
- `log_level`: 로그 레벨 ("debug", "info", "warn", "error")
- `metrics_interval`: 메트릭 출력 간격

## 출력 예시

```
[PSI] mem some avg10=2.45%
[PSI] cpu some avg10=1.23%
[PSI] io full avg10=0.87%
[NET] enp4s0 rx=1024000B/s tx=512000B/s
[PERF] MemBW total=1250.5MB/s (R=800.2 W=450.3)
[PERF] LLC mpki=15.67 hit=0.85 loads=125000 stores=75000
```

## 요구사항

- Linux 커널 4.20+ (PSI 지원)
- perf 도구 설치
- 네트워크 인터페이스 접근 권한
- `/proc/pressure` 및 `/sys/class/net` 접근 권한

## 라이선스

MIT License
