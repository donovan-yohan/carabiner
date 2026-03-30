package carabiner

import "time"

// Learning represents a quality pattern extracted from a gate failure.
type Learning struct {
	ID             string            `yaml:"id"`
	Created        time.Time         `yaml:"created"`
	Source         string            `yaml:"source"`
	Paths          []string          `yaml:"paths"`
	Tags           []string          `yaml:"tags,omitempty"`
	Pattern        string            `yaml:"pattern"`
	Recommendation string            `yaml:"recommendation"`
	RawInput       string            `yaml:"raw_input,omitempty"`
	Artifacts      LearningArtifacts `yaml:"artifacts,omitempty"`
}

type LearningArtifacts struct {
	Branch    string `yaml:"branch,omitempty"`
	Commit    string `yaml:"commit,omitempty"`
	DesignDoc string `yaml:"design_doc,omitempty"`
}

// Signal represents a single gate result for a learning.
type Signal struct {
	Timestamp  time.Time `json:"ts"`
	LearningID string    `json:"learning_id"`
	GateID     string    `json:"gate_id"`
	Result     string    `json:"result"`
	Branch     string    `json:"branch,omitempty"`
	Commit     string    `json:"commit,omitempty"`
	Files      []string  `json:"files,omitempty"`
}

// Config represents .carabiner/config.yaml.
type Config struct {
	Version   int             `yaml:"version"`
	Storage   StorageConfig   `yaml:"storage"`
	Quality   QualityConfig   `yaml:"quality"`
	Hierarchy HierarchyConfig `yaml:"hierarchy"`
}

type StorageConfig struct {
	Mode      string `yaml:"mode"`
	LocalPath string `yaml:"local_path,omitempty"`
}

type QualityConfig struct {
	Extraction ExtractionConfig `yaml:"extraction"`
	Check      CheckConfig      `yaml:"check"`
}

type ExtractionConfig struct {
	Model   string `yaml:"model"`
	Command string `yaml:"command"`
}

type CheckConfig struct {
	MaxResults     int  `yaml:"max_results"`
	IncludeDormant bool `yaml:"include_dormant"`
}

type HierarchyConfig struct {
	Parent string `yaml:"parent,omitempty"`
}

// GateResult is the normalized input format for --gate-result.
type GateResult struct {
	GateID    string   `json:"gate_id"`
	Result    string   `json:"result"`
	Score     int      `json:"score"`
	Rationale string   `json:"rationale"`
	Files     []string `json:"files"`
	Branch    string   `json:"branch"`
	Commit    string   `json:"commit"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig(mode string) Config {
	return Config{
		Version: 1,
		Storage: StorageConfig{Mode: mode},
		Quality: QualityConfig{
			Extraction: ExtractionConfig{
				Model:   "haiku",
				Command: "claude -p --model haiku",
			},
			Check: CheckConfig{
				MaxResults:     10,
				IncludeDormant: false,
			},
		},
	}
}
