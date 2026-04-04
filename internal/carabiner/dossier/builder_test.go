package dossier

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	_ "modernc.org/sqlite"
)

// setupIntegrationRepo creates a temp git repo with a file, a commit,
// and a git-ai note on that commit. Returns the repo dir and commit SHA.
func setupIntegrationRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s", args, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test User")

	// Write a file
	testFile := filepath.Join(dir, "main.go")
	content := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	run("git", "add", "main.go")
	run("git", "commit", "-m", "initial commit")

	// Get the commit SHA
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	shaBytes, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	sha := string(shaBytes[:len(shaBytes)-1]) // trim newline

	return dir, sha
}

func addGitAINote(t *testing.T, dir, sha, noteContent string) {
	t.Helper()
	cmd := exec.Command("git", "notes", "--ref", "ai", "add", "-m", noteContent, sha)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("adding note: %s", out)
	}
}

func setupAgentlyticsDB(t *testing.T, conversationID string) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "cache.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE chats (
			id TEXT PRIMARY KEY,
			source TEXT,
			folder TEXT,
			created_at TEXT,
			last_updated_at TEXT,
			name TEXT
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO chats (id, source, folder, created_at, last_updated_at, name)
		VALUES (?, ?, ?, ?, ?, ?)`,
		conversationID,
		"claude-code",
		"/test/project",
		"2026-03-28T14:32:00Z",
		"2026-03-28T15:47:00Z",
		"Implement main",
	)
	if err != nil {
		t.Fatal(err)
	}

	return dbPath
}

func TestBuild_FullChain(t *testing.T) {
	dir, sha := setupIntegrationRepo(t)
	conversationID := "655eb4a6-822f-4d52-8c06-cade4afdcd8d"

	// Add git-ai note with attestation covering lines 1-5
	noteContent := "main.go\n  abcdef1234567890 1-5\n---\n" +
		`{"schema_version":"3.0.0","base_commit_sha":"` + sha + `","prompts":{"abcdef1234567890":{"agent_id":{"tool":"claude_code","id":"` + conversationID + `","model":"claude-sonnet-4-5-20250514"},"messages":["write main"]}}}`
	addGitAINote(t, dir, sha, noteContent)

	// Set up agentlytics
	agentlyticsPath := setupAgentlyticsDB(t, conversationID)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	builder := NewBuilder(agentlyticsPath)
	d, err := builder.Build("main.go", 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.OverallConfidence != carabiner.ConfidenceHigh {
		t.Errorf("overall confidence = %q, want high", d.OverallConfidence)
	}
	if d.Blame == nil {
		t.Fatal("blame is nil")
	}
	if d.Blame.CommitSHA != sha {
		t.Errorf("blame SHA = %q, want %q", d.Blame.CommitSHA, sha)
	}
	if d.Session == nil {
		t.Fatal("session is nil")
	}
	if d.Session.ID != conversationID {
		t.Errorf("session ID = %q", d.Session.ID)
	}
	if d.Session.Tool != "claude_code" {
		t.Errorf("session tool = %q", d.Session.Tool)
	}
	if d.Session.Name != "Implement main" {
		t.Errorf("session name = %q", d.Session.Name)
	}

	// Should have 3 hops: line_to_commit, commit_to_session, session_to_transcript
	if len(d.Hops) != 3 {
		t.Fatalf("got %d hops, want 3", len(d.Hops))
	}
	for _, hop := range d.Hops {
		if hop.Confidence != carabiner.ConfidenceHigh {
			t.Errorf("hop %q confidence = %q, want high", hop.Name, hop.Confidence)
		}
	}
}

func TestBuild_NoGitAINote(t *testing.T) {
	dir, _ := setupIntegrationRepo(t)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	builder := NewBuilder("")
	d, err := builder.Build("main.go", 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.OverallConfidence != carabiner.ConfidenceMissing {
		t.Errorf("overall confidence = %q, want missing", d.OverallConfidence)
	}
	if d.Session != nil {
		t.Error("session should be nil when no git-ai note")
	}
	if d.Blame == nil {
		t.Error("blame should still be populated")
	}
}

func TestBuild_NoteButLineNotAttested(t *testing.T) {
	dir, sha := setupIntegrationRepo(t)

	// Attestation only covers line 1
	noteContent := "main.go\n  abcdef1234567890 1-1\n---\n" +
		`{"schema_version":"3.0.0","base_commit_sha":"` + sha + `","prompts":{"abcdef1234567890":{"agent_id":{"tool":"claude_code","id":"test-id","model":"test"},"messages":[]}}}`
	addGitAINote(t, dir, sha, noteContent)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	builder := NewBuilder("")
	// Query line 3, which isn't in the attested range
	d, err := builder.Build("main.go", 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.OverallConfidence != carabiner.ConfidenceMissing {
		t.Errorf("overall confidence = %q, want missing", d.OverallConfidence)
	}
}

func TestBuild_NoteButNoAgentlytics(t *testing.T) {
	dir, sha := setupIntegrationRepo(t)
	conversationID := "test-conv-id"

	noteContent := "main.go\n  abcdef1234567890 1-5\n---\n" +
		`{"schema_version":"3.0.0","base_commit_sha":"` + sha + `","prompts":{"abcdef1234567890":{"agent_id":{"tool":"claude_code","id":"` + conversationID + `","model":"test"},"messages":[]}}}`
	addGitAINote(t, dir, sha, noteContent)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	// No agentlytics path
	builder := NewBuilder("")
	d, err := builder.Build("main.go", 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have session from git-ai but no transcript hop
	if d.Session == nil {
		t.Fatal("session should be populated from git-ai note")
	}
	if d.Session.ID != conversationID {
		t.Errorf("session ID = %q", d.Session.ID)
	}

	// 2 hops: line_to_commit (high), commit_to_session (high)
	// No session_to_transcript hop since no agentlytics path
	if len(d.Hops) != 2 {
		t.Errorf("got %d hops, want 2", len(d.Hops))
	}
	if d.OverallConfidence != carabiner.ConfidenceHigh {
		t.Errorf("confidence = %q, want high (git-ai alone is sufficient)", d.OverallConfidence)
	}
}

func TestBuild_FileNotTracked(t *testing.T) {
	dir, _ := setupIntegrationRepo(t)

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	builder := NewBuilder("")
	_, err := builder.Build("nonexistent.go", 1, "")
	if err == nil {
		t.Error("expected error for untracked file")
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"./src/main.go", "src/main.go"},
		{"src/../src/main.go", "src/main.go"},
		{"src/main.go", "src/main.go"},
	}
	for _, tt := range tests {
		got := normalizePath(tt.input)
		if got != tt.want {
			t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
