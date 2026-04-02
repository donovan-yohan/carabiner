package enforce

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnforceConfig_Valid(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "carabiner-enforce-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write valid config
	configContent := `version: 1
tools:
  eslint:
    enabled: true
    command: npx
    args:
      - eslint
      - --max-warnings
      - "0"
    files:
      - "src/**/*.js"
      - "src/**/*.ts"
    timeout: 60
  golangci-lint:
    enabled: true
    command: golangci-lint
    args:
      - run
    files:
      - "**/*.go"
    timeout: 120
behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: true
`
	configPath := filepath.Join(tmpDir, "enforce.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	cfg, err := LoadEnforceConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadEnforceConfig returned error: %v", err)
	}

	// Verify loaded config
	if cfg.Version != 1 {
		t.Errorf("Expected version 1, got %d", cfg.Version)
	}

	if len(cfg.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(cfg.Tools))
	}

	eslint, ok := cfg.Tools["eslint"]
	if !ok {
		t.Fatal("Expected eslint tool to exist")
	}
	if !eslint.Enabled {
		t.Error("Expected eslint to be enabled")
	}
	if eslint.Command != "npx" {
		t.Errorf("Expected command 'npx', got %q", eslint.Command)
	}
	if len(eslint.Args) != 3 {
		t.Errorf("Expected 3 args for eslint, got %d", len(eslint.Args))
	}
	if eslint.Args[0] != "eslint" {
		t.Errorf("Expected first arg 'eslint', got %q", eslint.Args[0])
	}
	if eslint.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", eslint.Timeout)
	}

	if cfg.Behavior.FailOnWarning != true {
		t.Error("Expected FailOnWarning to be true")
	}
	if cfg.Behavior.StopOnFirstFail != false {
		t.Error("Expected StopOnFirstFail to be false")
	}
	if cfg.Behavior.Parallel != true {
		t.Error("Expected Parallel to be true")
	}
}

func TestLoadEnforceConfig_MissingFile_ReturnsError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-enforce-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg, err := LoadEnforceConfig(tmpDir)
	if err == nil {
		t.Fatal("Expected error when config file missing, got nil")
	}

	if cfg != nil {
		t.Error("Expected nil config when file missing")
	}
}

func TestLoadEnforceConfig_MalformedYAML_ReturnsError(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "carabiner-enforce-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write malformed YAML
	configContent := `version: 1
tools:
  eslint:
    enabled: true
    command: npx
      args:  # This is malformed - wrong indentation
      - eslint
`
	configPath := filepath.Join(tmpDir, "enforce.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config should return error
	_, err = LoadEnforceConfig(tmpDir)
	if err == nil {
		t.Fatal("Expected error for malformed YAML, got nil")
	}
}

func TestValidateEnforceConfig_Valid(t *testing.T) {
	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"eslint": {
				Enabled: true,
				Command: "npx",
				Args:    []string{"eslint", "--max-warnings", "0"},
				Files:   []string{"src/**/*.js"},
				Timeout: 60,
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   true,
			StopOnFirstFail: false,
			Parallel:        true,
		},
	}

	err := ValidateEnforceConfig(cfg)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
}

func TestValidateEnforceConfig_NoTools(t *testing.T) {
	cfg := &EnforceConfig{
		Version: 1,
		Tools:   map[string]ToolConfig{}, // Empty tools
		Behavior: BehaviorConfig{
			FailOnWarning:   true,
			StopOnFirstFail: false,
			Parallel:        true,
		},
	}

	err := ValidateEnforceConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for config with no enabled tools, got nil")
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestValidateEnforceConfig_UnknownVersion(t *testing.T) {
	cfg := &EnforceConfig{
		Version: 999, // Unknown version
		Tools: map[string]ToolConfig{
			"eslint": {
				Enabled: true,
				Command: "npx",
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning: true,
		},
	}

	err := ValidateEnforceConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for unknown version, got nil")
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestValidateEnforceConfig_MissingRequiredFields(t *testing.T) {
	// Tool with enabled=true but missing command
	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"eslint": {
				Enabled: true,
				Command: "", // Missing required field
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning: true,
		},
	}

	err := ValidateEnforceConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for missing command, got nil")
	}
}

func TestValidateEnforceConfig_DisabledToolsSkipped(t *testing.T) {
	// Tool with enabled=false should not require command
	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"eslint": {
				Enabled: false,
				Command: "", // Missing but OK since disabled
			},
			"prettier": {
				Enabled: true,
				Command: "npx",
				Args:    []string{"prettier", "--check"},
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning: true,
		},
	}

	err := ValidateEnforceConfig(cfg)
	if err != nil {
		t.Errorf("Expected valid config with disabled tool, got error: %v", err)
	}
}

func TestDefaultEnforceConfig(t *testing.T) {
	cfg := DefaultEnforceConfig()

	if cfg == nil {
		t.Fatal("Expected non-nil default config")
	}

	if cfg.Version != 1 {
		t.Errorf("Expected default version 1, got %d", cfg.Version)
	}

	if len(cfg.Tools) != 0 {
		t.Errorf("Expected 0 tools in default config, got %d", len(cfg.Tools))
	}

	// Verify default behavior settings
	if cfg.Behavior.FailOnWarning != true {
		t.Error("Expected default FailOnWarning to be true")
	}
	if cfg.Behavior.StopOnFirstFail != false {
		t.Error("Expected default StopOnFirstFail to be false")
	}
	if cfg.Behavior.Parallel != false {
		t.Error("Expected default Parallel to be false")
	}
}
