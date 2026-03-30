package carabiner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecord_SkipExtraction(t *testing.T) {
	tmp := t.TempDir()
	setupDirs(t, tmp)

	rawInput := "Auth route added without middleware registry update.\nThe /api/callback route is unprotected."
	learning, err := Record(tmp, "gate-1", rawInput, nil, true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if learning.ID == "" {
		t.Error("learning ID is empty")
	}
	if learning.Pattern != "Auth route added without middleware registry update." {
		t.Errorf("pattern = %q, want first line of raw input", learning.Pattern)
	}
	if learning.RawInput != rawInput {
		t.Error("raw_input not stored")
	}
	if learning.Source != "manual" {
		t.Errorf("source = %q, want 'manual'", learning.Source)
	}

	// Check learning file exists
	learningFile := filepath.Join(tmp, "quality", "learnings", learning.ID+".yaml")
	if _, err := os.Stat(learningFile); os.IsNotExist(err) {
		t.Error("learning YAML file not created")
	}

	// Check signal was appended
	signals, err := LoadSignals(tmp)
	if err != nil {
		t.Fatalf("loading signals: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("got %d signals, want 1", len(signals))
	}
	if signals[0].Result != "fail" {
		t.Errorf("signal result = %q, want 'fail'", signals[0].Result)
	}
	if signals[0].LearningID != learning.ID {
		t.Error("signal learning_id doesn't match")
	}
}

func TestRecord_WithGateResult(t *testing.T) {
	tmp := t.TempDir()
	setupDirs(t, tmp)

	gr := &GateResult{
		GateID:    "review-pr-42",
		Result:    "fail",
		Score:     3,
		Rationale: "Missing middleware registry update",
		Files:     []string{"src/auth/routes.ts", "src/auth/oauth.ts"},
		Branch:    "feature/oauth",
		Commit:    "abc1234",
	}
	rawInput := `{"gate_id":"review-pr-42","result":"fail","score":3,"rationale":"Missing middleware registry update","files":["src/auth/routes.ts","src/auth/oauth.ts"],"branch":"feature/oauth","commit":"abc1234"}`

	learning, err := Record(tmp, "review-pr-42", rawInput, gr, true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if learning.Source != "gate:review-pr-42" {
		t.Errorf("source = %q, want 'gate:review-pr-42'", learning.Source)
	}
	if learning.Artifacts.Branch != "feature/oauth" {
		t.Errorf("branch = %q, want 'feature/oauth'", learning.Artifacts.Branch)
	}
	if learning.Artifacts.Commit != "abc1234" {
		t.Errorf("commit = %q, want 'abc1234'", learning.Artifacts.Commit)
	}
	if len(learning.Paths) == 0 {
		t.Error("expected derived paths from gate result files")
	}

	// Verify paths are normalized with trailing slash
	for _, p := range learning.Paths {
		if p[len(p)-1] != '/' {
			t.Errorf("path %q missing trailing slash", p)
		}
	}
}

func TestRecord_MockExtraction(t *testing.T) {
	tmp := t.TempDir()
	setupDirs(t, tmp)

	// Use echo as a mock model CLI
	mockCmd := `echo {"pattern":"test pattern","recommendation":"test rec","paths":["src/"],"tags":["test"]}`

	rawInput := "Some gate failure rationale"
	learning, err := Record(tmp, "gate-1", rawInput, nil, false, mockCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if learning.Pattern != "test pattern" {
		t.Errorf("pattern = %q, want 'test pattern'", learning.Pattern)
	}
	if learning.Recommendation != "test rec" {
		t.Errorf("recommendation = %q, want 'test rec'", learning.Recommendation)
	}
	if learning.RawInput != rawInput {
		t.Error("raw_input not stored")
	}
}

func TestRecord_ExtractionFallback(t *testing.T) {
	tmp := t.TempDir()
	setupDirs(t, tmp)

	// Command that will fail
	mockCmd := "false"

	rawInput := "Some gate failure rationale"
	learning, err := Record(tmp, "gate-1", rawInput, nil, false, mockCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall back to raw input
	if learning.Pattern != "Some gate failure rationale" {
		t.Errorf("pattern = %q, want first line of raw input", learning.Pattern)
	}
}

func TestDerivePaths(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  int
	}{
		{"single file", []string{"src/auth/routes.ts"}, 1},
		{"same directory", []string{"src/auth/a.ts", "src/auth/b.ts"}, 1},
		{"different directories", []string{"src/auth/a.ts", "src/billing/b.ts"}, 2},
		{"root file", []string{"main.go"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := derivePaths(tt.files)
			if len(paths) != tt.want {
				t.Errorf("got %d paths, want %d: %v", len(paths), tt.want, paths)
			}
		})
	}
}

func TestNormalizePaths(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{[]string{"src/auth"}, []string{"src/auth/"}},
		{[]string{"src/auth/"}, []string{"src/auth/"}},
		{[]string{""}, []string{}},
		{[]string{"src/auth", "src/billing/"}, []string{"src/auth/", "src/billing/"}},
	}

	for _, tt := range tests {
		got := normalizePaths(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("normalizePaths(%v) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("normalizePaths(%v)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func setupDirs(t *testing.T, dir string) {
	t.Helper()
	for _, sub := range []string{
		"quality/learnings",
		"quality/signals",
	} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
}
