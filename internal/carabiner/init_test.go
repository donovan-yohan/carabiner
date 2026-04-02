package carabiner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit_RepoMode(t *testing.T) {
	tmp := t.TempDir()
	tmp, _ = filepath.EvalSymlinks(tmp) // macOS /var -> /private/var
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	dir, err := Init("repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmp, ".carabiner")
	if dir != expected {
		t.Errorf("dir = %q, want %q", dir, expected)
	}

	// Check directory structure
	for _, sub := range []string{
		"quality/learnings",
		"quality/signals",
	} {
		path := filepath.Join(expected, sub)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", sub)
		}
	}

	// Check config.yaml exists
	if _, err := os.Stat(filepath.Join(expected, "config.yaml")); os.IsNotExist(err) {
		t.Error("expected config.yaml to exist")
	}

	// Check .gitignore for signals
	gitignore := filepath.Join(expected, "quality", "signals", ".gitignore")
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		t.Error("expected signals .gitignore to exist")
	}
}

func TestInit_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	dir1, err := Init("repo")
	if err != nil {
		t.Fatalf("first init: %v", err)
	}

	// Write a marker file
	marker := filepath.Join(dir1, "marker.txt")
	if err := os.WriteFile(marker, []byte("test"), 0644); err != nil {
		t.Fatalf("write marker file: %v", err)
	}

	dir2, err := Init("repo")
	if err != nil {
		t.Fatalf("second init: %v", err)
	}

	if dir1 != dir2 {
		t.Errorf("dirs differ: %q vs %q", dir1, dir2)
	}

	// Marker should still exist (existing dir not modified)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		t.Error("marker file was removed, init modified existing directory")
	}
}

func TestInit_InvalidMode(t *testing.T) {
	_, err := Init("invalid")
	if err == nil {
		t.Error("expected error for invalid mode")
	}
}
