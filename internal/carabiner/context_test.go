package carabiner

import (
	"os/exec"
	"strings"
	"testing"
)

func TestSetGetWorkContext(t *testing.T) {
	_ = ClearWorkContext()

	ctx := NewWorkContext("TEST-123", "https://example.com/spec")

	if err := SetWorkContext(ctx); err != nil {
		t.Fatalf("SetWorkContext failed: %v", err)
	}

	retrieved, err := GetWorkContext()
	if err != nil {
		t.Fatalf("GetWorkContext failed: %v", err)
	}

	if retrieved.WorkItemRef != ctx.WorkItemRef {
		t.Errorf("WorkItemRef mismatch: got %q, want %q", retrieved.WorkItemRef, ctx.WorkItemRef)
	}
	if retrieved.SpecRef != ctx.SpecRef {
		t.Errorf("SpecRef mismatch: got %q, want %q", retrieved.SpecRef, ctx.SpecRef)
	}
	if retrieved.ContextBranch != ctx.ContextBranch {
		t.Errorf("ContextBranch mismatch: got %q, want %q", retrieved.ContextBranch, ctx.ContextBranch)
	}
	if retrieved.SetAt == "" {
		t.Error("SetAt should not be empty")
	}
	if retrieved.Source != "explicit" {
		t.Errorf("Source mismatch: got %q, want %q", retrieved.Source, "explicit")
	}

	_ = ClearWorkContext()
}

func TestClearWorkContext(t *testing.T) {
	ctx := NewWorkContext("TEST-456", "")
	if err := SetWorkContext(ctx); err != nil {
		t.Fatalf("SetWorkContext failed: %v", err)
	}

	result := ClearWorkContext()
	if !result.ClearSucceeded {
		t.Fatalf("ClearWorkContext failed to clear: %v", result.FailedKeys)
	}

	keys := []string{
		"carabiner.workItemRef",
		"carabiner.specRef",
		"carabiner.contextBranch",
		"carabiner.contextSetAt",
		"carabiner.contextSource",
	}

	for _, key := range keys {
		cmd := exec.Command("git", "config", "--worktree", "--get", key)
		out, _ := cmd.Output()
		if strings.TrimSpace(string(out)) != "" {
			t.Errorf("Key %q should be empty after clear", key)
		}
	}
}

func TestValidateWorkContext_MatchingBranch(t *testing.T) {
	_ = ClearWorkContext()

	ctx := NewWorkContext("TEST-789", "")

	if err := ValidateWorkContext(ctx); err != nil {
		t.Errorf("ValidateWorkContext failed on matching branch: %v", err)
	}

	_ = ClearWorkContext()
}

func TestValidateWorkContext_BranchMismatch(t *testing.T) {
	_ = ClearWorkContext()

	ctx := WorkContext{
		WorkItemRef:   "TEST-999",
		SpecRef:       "",
		ContextBranch: "nonexistent-branch-xyz",
		SetAt:         "2024-01-01T00:00:00Z",
		Source:        "explicit",
	}

	err := ValidateWorkContext(ctx)
	if err == nil {
		t.Error("ValidateWorkContext should fail on branch mismatch")
	}
	if !strings.Contains(err.Error(), "context branch mismatch") {
		t.Errorf("Expected branch mismatch error, got: %v", err)
	}
}

func TestValidateWorkContext_MissingWorkItem(t *testing.T) {
	_ = ClearWorkContext()

	ctx := WorkContext{
		WorkItemRef:   "",
		SpecRef:       "",
		ContextBranch: "main",
		SetAt:         "2024-01-01T00:00:00Z",
		Source:        "explicit",
	}

	err := ValidateWorkContext(ctx)
	if err == nil {
		t.Error("ValidateWorkContext should fail on missing work item")
	}
	if !strings.Contains(err.Error(), "work item ref is required") {
		t.Errorf("Expected work item required error, got: %v", err)
	}
}

func TestValidateWorkContext_MissingBranch(t *testing.T) {
	_ = ClearWorkContext()

	ctx := WorkContext{
		WorkItemRef:   "TEST-123",
		SpecRef:       "",
		ContextBranch: "",
		SetAt:         "2024-01-01T00:00:00Z",
		Source:        "explicit",
	}

	err := ValidateWorkContext(ctx)
	if err == nil {
		t.Error("ValidateWorkContext should fail on missing branch")
	}
	if !strings.Contains(err.Error(), "context branch is required") {
		t.Errorf("Expected branch required error, got: %v", err)
	}
}
