package config

import (
	"time"
)

// Config represents the main configuration structure
type Config struct {
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Output     OutputConfig     `yaml:"output"`
	PSIScope   PSIScopeConfig   `yaml:"psi_scope"`
}

// MonitoringConfig contains all monitoring-related settings
type MonitoringConfig struct {
	Network NetworkConfig `yaml:"network"`
	PSI     PSIConfig     `yaml:"psi"`
	Perf    PerfConfig    `yaml:"perf"`
}

// NetworkConfig contains network monitoring settings
type NetworkConfig struct {
	Interface string `yaml:"interface"`
	Interval  string `yaml:"interval"`
}

// PSIConfig contains PSI monitoring settings
type PSIConfig struct {
	Memory           PSIResourceConfig `yaml:"memory"`
	CPU              PSIResourceConfig `yaml:"cpu"`
	IO               PSIResourceConfig `yaml:"io"`
	MemoryPollInterval string          `yaml:"memory_poll_interval"`
}

// PSIResourceConfig contains settings for a specific PSI resource
type PSIResourceConfig struct {
	ThresholdUs int    `yaml:"threshold_us"`
	WindowUs    int    `yaml:"window_us"`
	Kind        string `yaml:"kind"`
}

// PerfConfig contains performance monitoring settings
type PerfConfig struct {
	Interval string   `yaml:"interval"`
	Events   []string `yaml:"events"`
}

// OutputConfig contains output-related settings
type OutputConfig struct {
	Console        bool   `yaml:"console"`
	LogLevel       string `yaml:"log_level"`
	MetricsInterval string `yaml:"metrics_interval"`
}

// PSIScopeConfig contains PSI scope settings
type PSIScopeConfig struct {
	Type       string `yaml:"type"`        // "system" or "cgroup"
	CgroupPath string `yaml:"cgroup_path"`
}

// Helper methods to convert string durations to time.Duration
func (c *Config) GetNetworkInterval() (time.Duration, error) {
	return time.ParseDuration(c.Monitoring.Network.Interval)
}

func (c *Config) GetPSIMemoryPollInterval() (time.Duration, error) {
	return time.ParseDuration(c.Monitoring.PSI.MemoryPollInterval)
}

func (c *Config) GetPerfInterval() (time.Duration, error) {
	return time.ParseDuration(c.Monitoring.Perf.Interval)
}

func (c *Config) GetMetricsInterval() (time.Duration, error) {
	return time.ParseDuration(c.Output.MetricsInterval)
}
