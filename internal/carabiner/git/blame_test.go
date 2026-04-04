package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test User"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s", c, out)
		}
	}

	// Create a file and commit it
	testFile := filepath.Join(dir, "hello.go")
	if err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmds = [][]string{
		{"git", "add", "hello.go"},
		{"git", "commit", "-m", "initial commit"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s", c, out)
		}
	}

	return dir
}

func TestBlame_HappyPath(t *testing.T) {
	dir := setupTestRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	result, err := Blame("hello.go", 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.File != "hello.go" {
		t.Errorf("file = %q, want %q", result.File, "hello.go")
	}
	if result.Line != 1 {
		t.Errorf("line = %d, want 1", result.Line)
	}
	if result.CommitSHA == "" {
		t.Error("commit SHA should not be empty")
	}
	if len(result.CommitSHA) != 40 {
		t.Errorf("commit SHA length = %d, want 40", len(result.CommitSHA))
	}
	if result.Author != "Test User" {
		t.Errorf("author = %q, want %q", result.Author, "Test User")
	}
	if result.Content != "package main" {
		t.Errorf("content = %q, want %q", result.Content, "package main")
	}
}

func TestBlame_FileNotTracked(t *testing.T) {
	dir := setupTestRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	_, err := Blame("nonexistent.go", 1, "")
	if err == nil {
		t.Error("expected error for untracked file")
	}
}

func TestBlame_LineOutOfRange(t *testing.T) {
	dir := setupTestRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	_, err := Blame("hello.go", 999, "")
	if err == nil {
		t.Error("expected error for out-of-range line")
	}
}

func TestBlame_WithRev(t *testing.T) {
	dir := setupTestRepo(t)
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	result, err := Blame("hello.go", 1, "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CommitSHA == "" {
		t.Error("commit SHA should not be empty")
	}
}

func TestParsePorcelainBlame(t *testing.T) {
	raw := `abcdef1234567890abcdef1234567890abcdef12 1 1 1
author Test User
author-mail <test@test.com>
author-time 1711800000
author-tz +0000
committer Test User
committer-mail <test@test.com>
committer-time 1711800000
committer-tz +0000
summary initial commit
filename hello.go
	package main`

	result, err := parsePorcelainBlame(raw, "hello.go", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CommitSHA != "abcdef1234567890abcdef1234567890abcdef12" {
		t.Errorf("commit = %q", result.CommitSHA)
	}
	if result.Author != "Test User" {
		t.Errorf("author = %q", result.Author)
	}
	if result.Content != "package main" {
		t.Errorf("content = %q", result.Content)
	}
	if result.Date != "1711800000" {
		t.Errorf("date = %q", result.Date)
	}
}
