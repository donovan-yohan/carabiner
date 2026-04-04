package gitai

import (
	"testing"
)

const sampleNote = `src/auth/handler.go
  a1b2c3d4e5f6g7h8 1-5,10-15
  ffffffffffffffff 20-25
src/utils/helpers.go
  a1b2c3d4e5f6g7h8 1-3
---
{"schema_version":"3.0.0","base_commit_sha":"abc123","prompts":{"a1b2c3d4e5f6g7h8":{"agent_id":{"tool":"claude_code","id":"655eb4a6-822f-4d52-8c06-cade4afdcd8d","model":"claude-sonnet-4-5-20250514"},"messages":["implement auth handler"]},"ffffffffffffffff":{"agent_id":{"tool":"opencode","id":"ses_abc123","model":"gpt-4o"},"messages":[]}}}`

func TestParseNote_Full(t *testing.T) {
	note, err := ParseNote(sampleNote)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(note.Attestations) != 2 {
		t.Fatalf("got %d attestations, want 2", len(note.Attestations))
	}

	// First file
	att := note.Attestations[0]
	if att.File != "src/auth/handler.go" {
		t.Errorf("file = %q", att.File)
	}
	if len(att.Sessions) != 2 {
		t.Fatalf("got %d sessions, want 2", len(att.Sessions))
	}
	if att.Sessions[0].Hash != "a1b2c3d4e5f6g7h8" {
		t.Errorf("hash = %q", att.Sessions[0].Hash)
	}
	if len(att.Sessions[0].LineRanges) != 2 {
		t.Fatalf("got %d ranges, want 2", len(att.Sessions[0].LineRanges))
	}
	if att.Sessions[0].LineRanges[0].Start != 1 || att.Sessions[0].LineRanges[0].End != 5 {
		t.Errorf("range = %+v", att.Sessions[0].LineRanges[0])
	}
	if att.Sessions[0].LineRanges[1].Start != 10 || att.Sessions[0].LineRanges[1].End != 15 {
		t.Errorf("range = %+v", att.Sessions[0].LineRanges[1])
	}

	// Second file
	att2 := note.Attestations[1]
	if att2.File != "src/utils/helpers.go" {
		t.Errorf("file = %q", att2.File)
	}

	// Metadata
	if note.Metadata.SchemaVersion != "3.0.0" {
		t.Errorf("schema_version = %q", note.Metadata.SchemaVersion)
	}
	if len(note.Metadata.Prompts) != 2 {
		t.Fatalf("got %d prompts, want 2", len(note.Metadata.Prompts))
	}

	prompt := note.Metadata.Prompts["a1b2c3d4e5f6g7h8"]
	if prompt.AgentID.Tool != "claude_code" {
		t.Errorf("tool = %q", prompt.AgentID.Tool)
	}
	if prompt.AgentID.ID != "655eb4a6-822f-4d52-8c06-cade4afdcd8d" {
		t.Errorf("id = %q", prompt.AgentID.ID)
	}
	if prompt.AgentID.Model != "claude-sonnet-4-5-20250514" {
		t.Errorf("model = %q", prompt.AgentID.Model)
	}
}

func TestParseNote_NoSeparator(t *testing.T) {
	_, err := ParseNote("just some text without separator")
	if err == nil {
		t.Error("expected error for missing separator")
	}
}

func TestParseNote_MalformedJSON(t *testing.T) {
	_, err := ParseNote("file.go\n  hash 1-5\n---\n{invalid json}")
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseNote_EmptyMessages(t *testing.T) {
	raw := `file.go
  abcdef1234567890 1-3
---
{"schema_version":"3.0.0","base_commit_sha":"abc","prompts":{"abcdef1234567890":{"agent_id":{"tool":"claude_code","id":"test-id","model":"test-model"},"messages":[]}}}`

	note, err := ParseNote(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	prompt := note.Metadata.Prompts["abcdef1234567890"]
	if len(prompt.Messages) != 0 {
		t.Errorf("expected empty messages, got %d", len(prompt.Messages))
	}
}

func TestFindSession_Found(t *testing.T) {
	note, err := ParseNote(sampleNote)
	if err != nil {
		t.Fatal(err)
	}

	hash, prompt, err := FindSession(note, "src/auth/handler.go", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "a1b2c3d4e5f6g7h8" {
		t.Errorf("hash = %q", hash)
	}
	if prompt == nil {
		t.Fatal("expected prompt, got nil")
	}
	if prompt.AgentID.ID != "655eb4a6-822f-4d52-8c06-cade4afdcd8d" {
		t.Errorf("id = %q", prompt.AgentID.ID)
	}
}

func TestFindSession_DifferentSession(t *testing.T) {
	note, err := ParseNote(sampleNote)
	if err != nil {
		t.Fatal(err)
	}

	hash, prompt, err := FindSession(note, "src/auth/handler.go", 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "ffffffffffffffff" {
		t.Errorf("hash = %q, want ffffffffffffffff", hash)
	}
	if prompt == nil {
		t.Fatal("expected prompt")
	}
	if prompt.AgentID.Tool != "opencode" {
		t.Errorf("tool = %q", prompt.AgentID.Tool)
	}
}

func TestFindSession_NotFound(t *testing.T) {
	note, err := ParseNote(sampleNote)
	if err != nil {
		t.Fatal(err)
	}

	hash, prompt, err := FindSession(note, "src/auth/handler.go", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "" {
		t.Errorf("expected empty hash, got %q", hash)
	}
	if prompt != nil {
		t.Error("expected nil prompt")
	}
}

func TestFindSession_WrongFile(t *testing.T) {
	note, err := ParseNote(sampleNote)
	if err != nil {
		t.Fatal(err)
	}

	hash, _, err := FindSession(note, "nonexistent.go", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "" {
		t.Errorf("expected empty hash, got %q", hash)
	}
}

func TestSessionHash(t *testing.T) {
	hash := SessionHash("claude_code", "655eb4a6-822f-4d52-8c06-cade4afdcd8d")
	if len(hash) != 16 {
		t.Errorf("hash length = %d, want 16", len(hash))
	}
	// Deterministic
	hash2 := SessionHash("claude_code", "655eb4a6-822f-4d52-8c06-cade4afdcd8d")
	if hash != hash2 {
		t.Errorf("non-deterministic: %q != %q", hash, hash2)
	}
}

func TestParseLineRanges_Single(t *testing.T) {
	ranges, err := parseLineRanges("5")
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 || ranges[0].Start != 5 || ranges[0].End != 5 {
		t.Errorf("got %+v", ranges)
	}
}

func TestParseLineRanges_Multiple(t *testing.T) {
	ranges, err := parseLineRanges("1-5,10-20,25")
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 3 {
		t.Fatalf("got %d ranges, want 3", len(ranges))
	}
	if ranges[2].Start != 25 || ranges[2].End != 25 {
		t.Errorf("single-line range = %+v", ranges[2])
	}
}
