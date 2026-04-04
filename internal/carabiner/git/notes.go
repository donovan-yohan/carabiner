package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsGitAIInstalled checks whether the git-ai binary is on PATH.
func IsGitAIInstalled() bool {
	_, err := exec.LookPath("git-ai")
	return err == nil
}

// ShowNote reads a git note from the given ref for the given commit.
// Returns the raw note text, or an error if no note exists.
func ShowNote(ref string, commitSHA string) (string, error) {
	out, err := RunGit("notes", "--ref", ref, "show", commitSHA)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "No note found") || strings.Contains(errStr, "no note found") {
			return "", fmt.Errorf("no git-ai note for commit %s", commitSHA)
		}
		return "", err
	}
	return out, nil
}

// HasNotesRef checks whether a git notes ref exists in this repo.
func HasNotesRef(ref string) bool {
	_, err := RunGit("rev-parse", "--verify", "refs/notes/"+ref)
	return err == nil
}
