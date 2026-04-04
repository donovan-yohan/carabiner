package git

import "testing"

func TestRunGit_Success(t *testing.T) {
	out, err := RunGit("rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "true" {
		t.Errorf("got %q, want %q", out, "true")
	}
}

func TestRunGit_BadCommand(t *testing.T) {
	_, err := RunGit("not-a-real-command")
	if err == nil {
		t.Error("expected error for invalid git command")
	}
}

func TestIsInsideWorkTree(t *testing.T) {
	if !IsInsideWorkTree() {
		t.Error("expected to be inside a git work tree")
	}
}
