package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowNote_NoNote(t *testing.T) {
	dir := setupTestRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	sha, err := RunGit("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ShowNote("ai", sha)
	if err == nil {
		t.Error("expected error for missing note")
	}
	if !strings.Contains(err.Error(), "no git-ai note") {
		t.Errorf("error should mention missing note, got: %v", err)
	}
}

func TestShowNote_WithNote(t *testing.T) {
	dir := setupTestRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	sha, err := RunGit("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	// Add a note
	cmd := exec.Command("git", "notes", "--ref", "ai", "add", "-m", "test note content", sha)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("adding note: %s", out)
	}

	note, err := ShowNote("ai", sha)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note != "test note content" {
		t.Errorf("note = %q, want %q", note, "test note content")
	}
}

func TestHasNotesRef(t *testing.T) {
	dir := setupTestRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	// No ai notes ref yet
	if HasNotesRef("ai") {
		t.Error("should not have ai notes ref in fresh repo")
	}

	// Create a note to establish the ref
	sha, _ := RunGit("rev-parse", "HEAD")
	cmd := exec.Command("git", "notes", "--ref", "ai", "add", "-m", "test", sha)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("adding note: %s", out)
	}

	if !HasNotesRef("ai") {
		t.Error("should have ai notes ref after adding a note")
	}
}

func TestHasNotesRef_Fresh(t *testing.T) {
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %s", out)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if HasNotesRef("ai") {
		t.Error("fresh repo should not have ai notes ref")
	}
}
