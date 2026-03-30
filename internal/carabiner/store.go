package carabiner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FindConfigDir walks up from CWD to find .carabiner/, or uses the override.
func FindConfigDir(override string) (string, error) {
	if override != "" {
		info, err := os.Stat(override)
		if err != nil {
			return "", fmt.Errorf("config dir %q not found: %w", override, err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("config dir %q is not a directory", override)
		}
		return override, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, ".carabiner")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .carabiner directory found (walked up to filesystem root)")
		}
		dir = parent
	}
}

// LoadConfig reads and parses config.yaml from the config directory.
func LoadConfig(configDir string) (*Config, error) {
	path := filepath.Join(configDir, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig("repo")
			return &cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config.yaml: %w", err)
	}
	return &cfg, nil
}

// LoadLearnings reads all YAML files from the learnings directory.
func LoadLearnings(configDir string) ([]Learning, error) {
	dir := filepath.Join(configDir, "quality", "learnings")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading learnings directory: %w", err)
	}

	var learnings []Learning
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", entry.Name(), err)
			continue
		}

		var l Learning
		if err := yaml.Unmarshal(data, &l); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping malformed %s: %v\n", entry.Name(), err)
			continue
		}
		learnings = append(learnings, l)
	}
	return learnings, nil
}

// SaveLearning atomically writes a learning YAML file.
func SaveLearning(configDir string, l *Learning) error {
	dir := filepath.Join(configDir, "quality", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating learnings directory: %w", err)
	}

	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshaling learning: %w", err)
	}

	target := filepath.Join(dir, l.ID+".yaml")
	tmp := target + ".tmp"

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// AppendSignal appends a single JSONL line to signals.jsonl.
func AppendSignal(configDir string, s *Signal) error {
	dir := filepath.Join(configDir, "quality", "signals")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating signals directory: %w", err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling signal: %w", err)
	}
	data = append(data, '\n')

	path := filepath.Join(dir, "signals.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening signals file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("appending signal: %w", err)
	}
	return nil
}

// LoadSignals reads all signals from the JSONL file.
func LoadSignals(configDir string) ([]Signal, error) {
	path := filepath.Join(configDir, "quality", "signals", "signals.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening signals file: %w", err)
	}
	defer f.Close()

	var signals []Signal
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var s Signal
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping malformed signal line: %v\n", err)
			continue
		}
		signals = append(signals, s)
	}
	return signals, scanner.Err()
}

// RepoSlug derives a stable identifier for the current repo.
func RepoSlug() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		return filepath.Base(strings.TrimSpace(string(out)))
	}
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(dir)
}
