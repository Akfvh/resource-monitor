package config

import (
	"strings"
	"strconv"
)

// Simple YAML parser for basic configuration
// This is a minimal implementation for demonstration purposes
// In production, consider using a proper YAML library like gopkg.in/yaml.v3

func parseYAML(data []byte, config *Config) error {
	lines := strings.Split(string(data), "\n")
	
	// Simple state machine for parsing
	var currentSection string
	var currentSubSection string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse sections
		if strings.HasSuffix(line, ":") {
			section := strings.TrimSuffix(line, ":")
			currentSection = section
			currentSubSection = ""
			continue
		}
		
		// Parse subsections
		if strings.Contains(line, ":") && !strings.HasPrefix(line, " ") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Remove quotes if present
				value = strings.Trim(value, "\"'")
				
				// Parse based on current section
				switch currentSection {
				case "monitoring":
					parseMonitoringConfig(key, value, config)
				case "output":
					parseOutputConfig(key, value, config)
				case "psi_scope":
					parsePSIScopeConfig(key, value, config)
				}
			}
		}
		
		// Parse indented subsections
		if strings.HasPrefix(line, "  ") && strings.HasSuffix(line, ":") {
			subsection := strings.TrimSpace(strings.TrimSuffix(line, ":"))
			currentSubSection = subsection
			continue
		}
		
		// Parse key-value pairs
		if strings.HasPrefix(line, "    ") && strings.Contains(line, ":") {
			parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"'")
				
				parseNestedConfig(currentSection, currentSubSection, key, value, config)
			}
		}
		
		// Parse list items
		if strings.HasPrefix(line, "      - ") {
			value := strings.TrimPrefix(line, "      - ")
			value = strings.Trim(value, "\"'")
			
			if currentSection == "monitoring" && currentSubSection == "perf" {
				config.Monitoring.Perf.Events = append(config.Monitoring.Perf.Events, value)
			}
		}
	}
	
	return nil
}

func parseMonitoringConfig(key, value string, config *Config) {
	switch key {
	case "network":
		// Network config is handled in nested parsing
	case "psi":
		// PSI config is handled in nested parsing
	case "perf":
		// Perf config is handled in nested parsing
	}
}

func parseOutputConfig(key, value string, config *Config) {
	switch key {
	case "console":
		config.Output.Console = value == "true"
	case "log_level":
		config.Output.LogLevel = value
	case "metrics_interval":
		config.Output.MetricsInterval = value
	}
}

func parsePSIScopeConfig(key, value string, config *Config) {
	switch key {
	case "type":
		config.PSIScope.Type = value
	case "cgroup_path":
		config.PSIScope.CgroupPath = value
	}
}

func parseNestedConfig(section, subsection, key, value string, config *Config) {
	if section == "monitoring" {
		switch subsection {
		case "network":
			switch key {
			case "interface":
				config.Monitoring.Network.Interface = value
			case "interval":
				config.Monitoring.Network.Interval = value
			}
		case "psi":
			// PSI configs are handled separately
		case "perf":
			switch key {
			case "interval":
				config.Monitoring.Perf.Interval = value
			}
		}
	}
}

// Parse PSI configuration from YAML
func parsePSIConfig(data []byte, config *Config) error {
	lines := strings.Split(string(data), "\n")
	
	var currentPSIResource string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Check for PSI resource sections
		if strings.HasPrefix(line, "    memory:") || strings.HasPrefix(line, "    cpu:") || strings.HasPrefix(line, "    io:") {
			currentPSIResource = strings.TrimSuffix(strings.TrimSpace(line), ":")
			continue
		}
		
		// Parse PSI resource configuration
		if strings.HasPrefix(line, "      ") && strings.Contains(line, ":") {
			parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"'")
				
				// Parse PSI resource config
				switch currentPSIResource {
				case "memory":
					parsePSIResourceConfig(key, value, &config.Monitoring.PSI.Memory)
				case "cpu":
					parsePSIResourceConfig(key, value, &config.Monitoring.PSI.CPU)
				case "io":
					parsePSIResourceConfig(key, value, &config.Monitoring.PSI.IO)
				}
			}
		}
		
		// Parse memory poll interval
		if strings.HasPrefix(line, "    memory_poll_interval:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"'")
				config.Monitoring.PSI.MemoryPollInterval = value
			}
		}
	}
	
	return nil
}

func parsePSIResourceConfig(key, value string, resource *PSIResourceConfig) {
	switch key {
	case "threshold_us":
		if val, err := strconv.Atoi(value); err == nil {
			resource.ThresholdUs = val
		}
	case "window_us":
		if val, err := strconv.Atoi(value); err == nil {
			resource.WindowUs = val
		}
	case "kind":
		resource.Kind = value
	}
}

// Enhanced YAML parser that handles the full config
func parseYAMLEnhanced(data []byte, config *Config) error {
	// First pass: basic parsing
	if err := parseYAML(data, config); err != nil {
		return err
	}
	
	// Second pass: PSI-specific parsing
	if err := parsePSIConfig(data, config); err != nil {
		return err
	}
	
	return nil
}
