package enforce

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadEnforceConfig reads and parses enforce.yaml from the config directory.
// Returns an error if the file doesn't exist.
func LoadEnforceConfig(configDir string) (*EnforceConfig, error) {
	path := filepath.Join(configDir, "enforce.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("enforce config not found: %s", path)
		}
		return nil, fmt.Errorf("reading enforce config: %w", err)
	}

	var cfg EnforceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing enforce.yaml: %w", err)
	}

	return &cfg, nil
}

// DefaultEnforceConfig returns a config with sensible defaults.
func DefaultEnforceConfig() *EnforceConfig {
	return &EnforceConfig{
		Version: 1,
		Tools:   make(map[string]ToolConfig),
		Behavior: BehaviorConfig{
			FailOnWarning:   true,
			StopOnFirstFail: false,
			Parallel:        false,
		},
	}
}

// ValidateEnforceConfig validates the enforce configuration.
// Returns an error if the config is invalid.
func ValidateEnforceConfig(cfg *EnforceConfig) error {
	if cfg.Version != 1 {
		return fmt.Errorf("unknown config version %d: only version 1 is supported", cfg.Version)
	}

	hasEnabledTool := false
	for name, tool := range cfg.Tools {
		if tool.Enabled {
			hasEnabledTool = true
			if tool.Command == "" {
				return fmt.Errorf("tool %q is enabled but missing required field 'command'", name)
			}
		}
	}

	if !hasEnabledTool {
		return fmt.Errorf("at least one tool must be enabled")
	}

	return nil
}
