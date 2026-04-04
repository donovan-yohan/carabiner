package carabiner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindConfigDir_WalkUp(t *testing.T) {
	tmp := t.TempDir()
	// Resolve symlinks (macOS /var -> /private/var)
	tmp, _ = filepath.EvalSymlinks(tmp)
	carabinerDir := filepath.Join(tmp, ".carabiner")
	if err := os.Mkdir(carabinerDir, 0755); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(subdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	got, err := FindConfigDir("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != carabinerDir {
		t.Errorf("got %q, want %q", got, carabinerDir)
	}
}

func TestFindConfigDir_Override(t *testing.T) {
	tmp := t.TempDir()
	got, err := FindConfigDir(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tmp {
		t.Errorf("got %q, want %q", got, tmp)
	}
}

func TestFindConfigDir_NotFound(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	_, err := FindConfigDir("")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestFindConfigDir_OverrideNotExist(t *testing.T) {
	_, err := FindConfigDir("/nonexistent/path")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
