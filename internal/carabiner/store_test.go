package carabiner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	defer os.Chdir(orig)
	os.Chdir(subdir)

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
	defer os.Chdir(orig)
	os.Chdir(tmp)

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

func TestLoadConfig_Default(t *testing.T) {
	tmp := t.TempDir()
	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("version = %d, want 1", cfg.Version)
	}
	if cfg.Storage.Mode != "repo" {
		t.Errorf("mode = %q, want 'repo'", cfg.Storage.Mode)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	tmp := t.TempDir()
	content := "version: 2\nstorage:\n  mode: local\n"
	if err := os.WriteFile(filepath.Join(tmp, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 2 {
		t.Errorf("version = %d, want 2", cfg.Version)
	}
	if cfg.Storage.Mode != "local" {
		t.Errorf("mode = %q, want 'local'", cfg.Storage.Mode)
	}
}

func TestSaveAndLoadLearning(t *testing.T) {
	tmp := t.TempDir()
	learning := &Learning{
		ID:             "test-uuid-1",
		Created:        time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC),
		Source:         "gate:test",
		Paths:          []string{"src/auth/"},
		Tags:           []string{"auth"},
		Pattern:        "test pattern",
		Recommendation: "test recommendation",
	}

	if err := SaveLearning(tmp, learning); err != nil {
		t.Fatalf("save: %v", err)
	}

	learnings, err := LoadLearnings(tmp)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(learnings) != 1 {
		t.Fatalf("got %d learnings, want 1", len(learnings))
	}
	if learnings[0].ID != "test-uuid-1" {
		t.Errorf("id = %q, want 'test-uuid-1'", learnings[0].ID)
	}
	if learnings[0].Pattern != "test pattern" {
		t.Errorf("pattern = %q, want 'test pattern'", learnings[0].Pattern)
	}
}

func TestLoadLearnings_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	learnings, err := LoadLearnings(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(learnings) != 0 {
		t.Errorf("got %d learnings, want 0", len(learnings))
	}
}

func TestLoadLearnings_MalformedYAML(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "quality", "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte(":::invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	learnings, err := LoadLearnings(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(learnings) != 0 {
		t.Errorf("got %d learnings, want 0 (malformed should be skipped)", len(learnings))
	}
}

func TestAppendAndLoadSignals(t *testing.T) {
	tmp := t.TempDir()
	now := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)

	s1 := &Signal{Timestamp: now, LearningID: "id-1", GateID: "gate-1", Result: "fail"}
	s2 := &Signal{Timestamp: now, LearningID: "id-1", GateID: "gate-2", Result: "pass"}

	if err := AppendSignal(tmp, s1); err != nil {
		t.Fatalf("append s1: %v", err)
	}
	if err := AppendSignal(tmp, s2); err != nil {
		t.Fatalf("append s2: %v", err)
	}

	signals, err := LoadSignals(tmp)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(signals) != 2 {
		t.Fatalf("got %d signals, want 2", len(signals))
	}
	if signals[0].Result != "fail" {
		t.Errorf("s1 result = %q, want 'fail'", signals[0].Result)
	}
	if signals[1].Result != "pass" {
		t.Errorf("s2 result = %q, want 'pass'", signals[1].Result)
	}
}

func TestLoadSignals_Empty(t *testing.T) {
	tmp := t.TempDir()
	signals, err := LoadSignals(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("got %d signals, want 0", len(signals))
	}
}
