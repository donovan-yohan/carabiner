package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunGit executes a git command and returns its trimmed stdout.
// Unlike the old gitValue(), this returns errors instead of swallowing them.
func RunGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			return "", fmt.Errorf("git %s: %s", args[0], stderr)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsInsideWorkTree returns true if the current directory is inside a git repo.
func IsInsideWorkTree() bool {
	_, err := RunGit("rev-parse", "--is-inside-work-tree")
	return err == nil
}
