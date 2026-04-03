package enforce

// EnforceConfig represents the configuration for the enforce command.
type EnforceConfig struct {
	Version          int                   `yaml:"version"`
	Tools            map[string]ToolConfig `yaml:"tools"`
	Behavior         BehaviorConfig        `yaml:"behavior"`
	AgentValidations []ValidationConfig    `yaml:"agent_validations"`
}

// ValidationConfig represents a single agent validation.
type ValidationConfig struct {
	Name   string   `yaml:"name"`
	Script string   `yaml:"script"`
	Files  []string `yaml:"files"`
}

// ToolConfig represents configuration for a single enforcement tool.
type ToolConfig struct {
	Enabled bool     `yaml:"enabled"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Files   []string `yaml:"files"`
	Timeout int      `yaml:"timeout"` // seconds, 0 = no timeout
}

// BehaviorConfig represents global behavior settings for enforcement.
type BehaviorConfig struct {
	FailOnWarning   bool `yaml:"fail_on_warning"`
	StopOnFirstFail bool `yaml:"stop_on_first_failure"`
	Parallel        bool `yaml:"parallel"`
}
