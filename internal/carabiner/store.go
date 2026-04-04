package carabiner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
