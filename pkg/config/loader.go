package config

import (
	"fmt"
	"os"
	"path/filepath"
	
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from YAML file
func LoadConfig(configPath string) (*Config, error) {
	// If no path provided, try default locations
	if configPath == "" {
		configPath = findConfigFile()
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// findConfigFile searches for config file in common locations
func findConfigFile() string {
	// Search in current directory and parent directories
	dir, _ := os.Getwd()
	for {
		configPath := filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	
	// Fallback to current directory
	return "config.yaml"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate PSI scope
	if c.PSIScope.Type != "system" && c.PSIScope.Type != "cgroup" {
		return fmt.Errorf("invalid PSI scope type: %s (must be 'system' or 'cgroup')", c.PSIScope.Type)
	}

	// Validate PSI resource kinds
	psiResources := []PSIResourceConfig{c.Monitoring.PSI.Memory, c.Monitoring.PSI.CPU, c.Monitoring.PSI.IO}
	for i, resource := range psiResources {
		resourceNames := []string{"memory", "cpu", "io"}
		if resource.Kind != "some" && resource.Kind != "full" {
			return fmt.Errorf("invalid PSI %s kind: %s (must be 'some' or 'full')", resourceNames[i], resource.Kind)
		}
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if c.Output.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.Output.LogLevel)
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Monitoring: MonitoringConfig{
			Network: NetworkConfig{
				Interface: "enp4s0",
				Interval:  "1s",
			},
			PSI: PSIConfig{
				Memory: PSIResourceConfig{
					ThresholdUs: 150000,
					WindowUs:    1000000,
					Kind:        "some",
				},
				CPU: PSIResourceConfig{
					ThresholdUs: 100000,
					WindowUs:    1000000,
					Kind:        "some",
				},
				IO: PSIResourceConfig{
					ThresholdUs: 150000,
					WindowUs:    1000000,
					Kind:        "full",
				},
				MemoryPollInterval: "1s",
			},
			Perf: PerfConfig{
				Interval: "1s",
				Events: []string{
					"LLC-loads",
					"LLC-load-misses",
					"LLC-stores",
					"LLC-store-misses",
					"instructions",
					"unc_m_cas_count_rd",
					"unc_m_cas_count_wr",
				},
			},
		},
		Output: OutputConfig{
			Console:        true,
			LogLevel:       "info",
			MetricsInterval: "1s",
		},
		PSIScope: PSIScopeConfig{
			Type:       "system",
			CgroupPath: "/sys/fs/cgroup",
		},
	}
}
