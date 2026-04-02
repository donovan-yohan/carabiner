package carabiner

import (
	"os"
	"strings"
	"testing"
	"time"
)

func setupLearnings(t *testing.T, learnings []Learning) string {
	t.Helper()
	tmp := t.TempDir()
	for _, l := range learnings {
		if err := SaveLearning(tmp, &l); err != nil {
			t.Fatal(err)
		}
	}
	return tmp
}

func TestCheck_PathPrefixMatch(t *testing.T) {
	tests := []struct {
		name          string
		learningPaths []string
		inputFiles    []string
		wantMatch     bool
	}{
		{
			name:          "exact directory match",
			learningPaths: []string{"src/auth/"},
			inputFiles:    []string{"src/auth/routes.ts"},
			wantMatch:     true,
		},
		{
			name:          "non-match similar prefix",
			learningPaths: []string{"src/auth/"},
			inputFiles:    []string{"src/authentication/routes.ts"},
			wantMatch:     false,
		},
		{
			name:          "trailing slash normalization",
			learningPaths: []string{"src/auth"},
			inputFiles:    []string{"src/auth/routes.ts"},
			wantMatch:     true,
		},
		{
			name:          "no trailing slash non-match",
			learningPaths: []string{"src/auth"},
			inputFiles:    []string{"src/authentication/routes.ts"},
			wantMatch:     false,
		},
		{
			name:          "multiple learning paths, one matches",
			learningPaths: []string{"src/auth/", "src/middleware/"},
			inputFiles:    []string{"src/middleware/cors.ts"},
			wantMatch:     true,
		},
		{
			name:          "multiple input files, one matches",
			learningPaths: []string{"src/auth/"},
			inputFiles:    []string{"src/billing/pay.ts", "src/auth/login.ts"},
			wantMatch:     true,
		},
		{
			name:          "no match at all",
			learningPaths: []string{"src/auth/"},
			inputFiles:    []string{"src/billing/pay.ts"},
			wantMatch:     false,
		},
		{
			name:          "deep nesting match",
			learningPaths: []string{"src/"},
			inputFiles:    []string{"src/auth/oauth/callback.ts"},
			wantMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesAnyPath(tt.learningPaths, tt.inputFiles)
			if got != tt.wantMatch {
				t.Errorf("matchesAnyPath(%v, %v) = %v, want %v",
					tt.learningPaths, tt.inputFiles, got, tt.wantMatch)
			}
		})
	}
}

func TestCheck_ReturnsMatchedLearnings(t *testing.T) {
	now := time.Now().UTC()
	learnings := []Learning{
		{ID: "1", Created: now, Paths: []string{"src/auth/"}, Pattern: "auth pattern"},
		{ID: "2", Created: now, Paths: []string{"src/billing/"}, Pattern: "billing pattern"},
		{ID: "3", Created: now, Paths: []string{"src/auth/", "src/middleware/"}, Pattern: "multi-path pattern"},
	}
	tmp := setupLearnings(t, learnings)

	matched, err := Check(tmp, []string{"src/auth/routes.ts"}, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matched) != 2 {
		t.Fatalf("got %d matches, want 2", len(matched))
	}

	patterns := make(map[string]bool)
	for _, m := range matched {
		patterns[m.Pattern] = true
	}
	if !patterns["auth pattern"] {
		t.Error("missing 'auth pattern'")
	}
	if !patterns["multi-path pattern"] {
		t.Error("missing 'multi-path pattern'")
	}
}

func TestCheck_SortsByCreatedDescending(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	learnings := []Learning{
		{ID: "old", Created: t1, Paths: []string{"src/"}, Pattern: "oldest"},
		{ID: "new", Created: t3, Paths: []string{"src/"}, Pattern: "newest"},
		{ID: "mid", Created: t2, Paths: []string{"src/"}, Pattern: "middle"},
	}
	tmp := setupLearnings(t, learnings)

	matched, err := Check(tmp, []string{"src/file.go"}, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matched) != 3 {
		t.Fatalf("got %d matches, want 3", len(matched))
	}
	if matched[0].Pattern != "newest" {
		t.Errorf("first = %q, want 'newest'", matched[0].Pattern)
	}
	if matched[2].Pattern != "oldest" {
		t.Errorf("last = %q, want 'oldest'", matched[2].Pattern)
	}
}

func TestCheck_MaxResults(t *testing.T) {
	now := time.Now().UTC()
	var learnings []Learning
	for i := 0; i < 5; i++ {
		learnings = append(learnings, Learning{
			ID:      string(rune('a' + i)),
			Created: now,
			Paths:   []string{"src/"},
			Pattern: "pattern",
		})
	}
	tmp := setupLearnings(t, learnings)

	matched, err := Check(tmp, []string{"src/file.go"}, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matched) != 3 {
		t.Errorf("got %d matches, want 3", len(matched))
	}
}

func TestCheck_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	matched, err := Check(tmp, []string{"src/file.go"}, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matched) != 0 {
		t.Errorf("got %d matches, want 0", len(matched))
	}
}

func TestFormatMarkdown(t *testing.T) {
	learnings := []Learning{
		{
			Pattern:        "Auth route changes need middleware updates",
			Recommendation: "Update the middleware registry after changing auth routes",
			Source:         "gate:silent-failure-hunter",
			Tags:           []string{"auth", "middleware"},
		},
	}

	output := FormatMarkdown(learnings)

	if !strings.Contains(output, "## Carabiner Quality Patterns (1 relevant)") {
		t.Error("missing header")
	}
	if !strings.Contains(output, "### Auth route changes need middleware updates") {
		t.Error("missing pattern")
	}
	if !strings.Contains(output, "**Recommendation:**") {
		t.Error("missing recommendation")
	}
	if !strings.Contains(output, "**Tags:** auth, middleware") {
		t.Error("missing tags")
	}
}

func TestFormatMarkdown_Empty(t *testing.T) {
	output := FormatMarkdown(nil)
	if output != "" {
		t.Errorf("expected empty string, got %q", output)
	}
}

func TestCheck_NoMatchReturnsEmpty(t *testing.T) {
	now := time.Now().UTC()
	learnings := []Learning{
		{ID: "1", Created: now, Paths: []string{"src/auth/"}, Pattern: "auth only"},
	}
	tmp := setupLearnings(t, learnings)

	matched, err := Check(tmp, []string{"src/billing/pay.ts"}, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matched) != 0 {
		t.Errorf("got %d matches, want 0", len(matched))
	}
}

// Verify that os.Chdir test cleanup works
func init() {
	os.Setenv("CARABINER_TEST", "1")
}
