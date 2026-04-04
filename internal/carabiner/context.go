package carabiner

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type WorkContext struct {
	WorkItemRef   string
	SpecRef       string
	ContextBranch string
	SetAt         string
	Source        string
}

func SetWorkContext(ctx WorkContext) error {
	if err := ValidateWorkContext(ctx); err != nil {
		return err
	}

	keys := map[string]string{
		"carabiner.workItemRef":   ctx.WorkItemRef,
		"carabiner.specRef":       ctx.SpecRef,
		"carabiner.contextBranch": ctx.ContextBranch,
		"carabiner.contextSetAt":  ctx.SetAt,
		"carabiner.contextSource": ctx.Source,
	}

	for key, value := range keys {
		cmd := exec.Command("git", "config", "--worktree", key, value)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}

	return nil
}

func GetWorkContext() (WorkContext, error) {
	ctx := WorkContext{}

	ctx.WorkItemRef = gitConfigValue("carabiner.workItemRef")
	ctx.SpecRef = gitConfigValue("carabiner.specRef")
	ctx.ContextBranch = gitConfigValue("carabiner.contextBranch")
	ctx.SetAt = gitConfigValue("carabiner.contextSetAt")
	ctx.Source = gitConfigValue("carabiner.contextSource")

	return ctx, nil
}

// ClearResult holds the outcome of a ClearWorkContext operation.
type ClearResult struct {
	// FailedKeys lists keys that could not be unset. If non-empty, the context
	// may still contain stale values for these keys.
	FailedKeys []string
	// ClearSucceeded is true only if all keys were successfully unset.
	ClearSucceeded bool
}

// ClearWorkContext removes all carabiner context keys from git worktree config.
// It continues even if individual keys fail to unset (e.g., already absent).
// Callers should check the returned ClearResult to report partial failures.
func ClearWorkContext() ClearResult {
	keys := []string{
		"carabiner.workItemRef",
		"carabiner.specRef",
		"carabiner.contextBranch",
		"carabiner.contextSetAt",
		"carabiner.contextSource",
	}

	result := ClearResult{ClearSucceeded: true}
	for _, key := range keys {
		cmd := exec.Command("git", "config", "--worktree", "--unset", key)
		if err := cmd.Run(); err != nil {
			result.FailedKeys = append(result.FailedKeys, key)
			result.ClearSucceeded = false
		}
	}

	return result
}

func CurrentBranch() string {
	return gitConfigValue("carabiner.contextBranch")
}

func ValidateWorkContext(ctx WorkContext) error {
	if ctx.WorkItemRef == "" {
		return fmt.Errorf("work item ref is required")
	}

	if ctx.ContextBranch == "" {
		return fmt.Errorf("context branch is required")
	}

	currentBranch := gitValue("rev-parse", "--abbrev-ref", "HEAD")
	if currentBranch != "" && ctx.ContextBranch != currentBranch {
		return fmt.Errorf("context branch mismatch: current=%s stored=%s", currentBranch, ctx.ContextBranch)
	}

	return nil
}

func gitConfigValue(key string) string {
	cmd := exec.Command("git", "config", "--worktree", "--get", key)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitValue(args ...string) string {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func NewWorkContext(workItemRef, specRef string) WorkContext {
	return WorkContext{
		WorkItemRef:   workItemRef,
		SpecRef:       specRef,
		ContextBranch: gitValue("rev-parse", "--abbrev-ref", "HEAD"),
		SetAt:         time.Now().Format(time.RFC3339),
		Source:        "explicit",
	}
}
