package events

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkflowTablesExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("initializing DB: %v", err)
	}
	defer db.Close()

	tables := []string{"work_context_events", "workflow_events", "git_attributions"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s does not exist: %v", table, err)
		}
		if name != table {
			t.Errorf("expected table %s, got %s", table, name)
		}
	}

	indexes := []string{
		"idx_workflow_timestamp",
		"idx_workflow_workflow",
		"idx_workflow_branch",
		"idx_workflow_session",
		"idx_workflow_failure",
		"idx_context_timestamp",
	}
	for _, index := range indexes {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&name)
		if err != nil {
			t.Errorf("index %s does not exist: %v", index, err)
		}
		if name != index {
			t.Errorf("expected index %s, got %s", index, name)
		}
	}
}

func TestAppendWorkContextEvent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("initializing DB: %v", err)
	}
	defer db.Close()

	event := &WorkContextEvent{
		ID:          "test-id-1",
		Timestamp:   time.Now().UTC(),
		WorkItemRef: "issue-123",
		SpecRef:     "spec-456",
		Branch:      "feature/test",
		Source:      "manual",
		Metadata:    `{"key": "value"}`,
	}

	err = AppendWorkContextEvent(db, event)
	if err != nil {
		t.Fatalf("appending work context event: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM work_context_events WHERE id = ?", event.ID).Scan(&count)
	if err != nil {
		t.Fatalf("querying work context events: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
}

func TestAppendWorkflowEvent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("initializing DB: %v", err)
	}
	defer db.Close()

	event := &WorkflowEvent{
		ID:                "workflow-1",
		Timestamp:         time.Now().UTC(),
		Workflow:          "test-workflow",
		EventType:         "start",
		ExternalSessionID: "session-123",
		ExternalRunID:     "run-456",
		RepoPath:          "/path/to/repo",
		Branch:            "main",
		CommitSHA:         "abc123",
		Agent:             "test-agent",
		Model:             "test-model",
		DurationMs:        1000,
		FailureCategory:   "",
		Metadata:          `{"key": "value"}`,
	}

	err = AppendWorkflowEvent(db, event)
	if err != nil {
		t.Fatalf("appending workflow event: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM workflow_events WHERE id = ?", event.ID).Scan(&count)
	if err != nil {
		t.Fatalf("querying workflow events: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
}

func TestUpsertGitAttribution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("initializing DB: %v", err)
	}
	defer db.Close()

	attr1 := &GitAttribution{
		CommitSHA:      "abc123",
		WorkItemRef:    "issue-123",
		SpecRef:        "spec-456",
		Branch:         "feature/test",
		TrailerPayload: `{"type": "work-item"}`,
		CreatedAt:      time.Now().UTC(),
	}

	err = UpsertGitAttribution(db, attr1)
	if err != nil {
		t.Fatalf("upserting git attribution: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM git_attributions WHERE commit_sha = ?", attr1.CommitSHA).Scan(&count)
	if err != nil {
		t.Fatalf("querying git attributions: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 attribution, got %d", count)
	}

	attr2 := &GitAttribution{
		CommitSHA:      "abc123",
		WorkItemRef:    "issue-789",
		SpecRef:        "spec-999",
		Branch:         "feature/updated",
		TrailerPayload: `{"type": "updated"}`,
		CreatedAt:      time.Now().UTC(),
	}

	err = UpsertGitAttribution(db, attr2)
	if err != nil {
		t.Fatalf("upserting git attribution (update): %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM git_attributions WHERE commit_sha = ?", attr1.CommitSHA).Scan(&count)
	if err != nil {
		t.Fatalf("querying git attributions after upsert: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 attribution after upsert, got %d", count)
	}

	var workItemRef string
	err = db.QueryRow("SELECT work_item_ref FROM git_attributions WHERE commit_sha = ?", attr1.CommitSHA).Scan(&workItemRef)
	if err != nil {
		t.Fatalf("querying work_item_ref: %v", err)
	}
	if workItemRef != attr2.WorkItemRef {
		t.Errorf("expected work_item_ref %s, got %s", attr2.WorkItemRef, workItemRef)
	}
}

func TestListWorkflowEvents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("initializing DB: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	events := []*WorkflowEvent{
		{
			ID:        "wf-1",
			Timestamp: now.Add(-2 * time.Hour),
			Workflow:  "workflow-a",
			EventType: "start",
		},
		{
			ID:        "wf-2",
			Timestamp: now.Add(-1 * time.Hour),
			Workflow:  "workflow-b",
			EventType: "start",
		},
		{
			ID:        "wf-3",
			Timestamp: now,
			Workflow:  "workflow-a",
			EventType: "end",
		},
	}

	for _, e := range events {
		err = AppendWorkflowEvent(db, e)
		if err != nil {
			t.Fatalf("appending workflow event: %v", err)
		}
	}

	filtered, err := ListWorkflowEvents(db, "workflow-a", 0)
	if err != nil {
		t.Fatalf("listing workflow events: %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("expected 2 events for workflow-a, got %d", len(filtered))
	}

	all, err := ListWorkflowEvents(db, "", 0)
	if err != nil {
		t.Fatalf("listing all workflow events: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 events, got %d", len(all))
	}
	if all[0].ID != "wf-3" {
		t.Errorf("expected first event to be wf-3, got %s", all[0].ID)
	}

	limited, err := ListWorkflowEvents(db, "", 2)
	if err != nil {
		t.Fatalf("listing workflow events with limit: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 events with limit, got %d", len(limited))
	}
}

func TestListRecentAttributions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("initializing DB: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	attributions := []*GitAttribution{
		{
			CommitSHA:      "commit-1",
			WorkItemRef:    "issue-1",
			Branch:         "branch-1",
			TrailerPayload: "{}",
			CreatedAt:      now.Add(-2 * time.Hour),
		},
		{
			CommitSHA:      "commit-2",
			WorkItemRef:    "issue-2",
			Branch:         "branch-2",
			TrailerPayload: "{}",
			CreatedAt:      now.Add(-1 * time.Hour),
		},
		{
			CommitSHA:      "commit-3",
			WorkItemRef:    "issue-3",
			Branch:         "branch-3",
			TrailerPayload: "{}",
			CreatedAt:      now,
		},
	}

	for _, a := range attributions {
		err = UpsertGitAttribution(db, a)
		if err != nil {
			t.Fatalf("upserting git attribution: %v", err)
		}
	}

	all, err := ListRecentAttributions(db, 0)
	if err != nil {
		t.Fatalf("listing recent attributions: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 attributions, got %d", len(all))
	}
	if all[0].CommitSHA != "commit-3" {
		t.Errorf("expected first attribution to be commit-3, got %s", all[0].CommitSHA)
	}

	limited, err := ListRecentAttributions(db, 2)
	if err != nil {
		t.Fatalf("listing recent attributions with limit: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 attributions with limit, got %d", len(limited))
	}
}

func TestListWorkContextEvents(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "carabiner-test-*")
	if err != nil {
		t.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("initializing DB: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	events := []*WorkContextEvent{
		{
			ID:          "ctx-1",
			Timestamp:   now.Add(-2 * time.Hour),
			WorkItemRef: "issue-1",
			Branch:      "branch-1",
			Source:      "source-1",
		},
		{
			ID:          "ctx-2",
			Timestamp:   now.Add(-1 * time.Hour),
			WorkItemRef: "issue-2",
			Branch:      "branch-2",
			Source:      "source-2",
		},
		{
			ID:          "ctx-3",
			Timestamp:   now,
			WorkItemRef: "issue-3",
			Branch:      "branch-3",
			Source:      "source-3",
		},
	}

	for _, e := range events {
		err = AppendWorkContextEvent(db, e)
		if err != nil {
			t.Fatalf("appending work context event: %v", err)
		}
	}

	all, err := ListWorkContextEvents(db, 0)
	if err != nil {
		t.Fatalf("listing work context events: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 events, got %d", len(all))
	}
	if all[0].ID != "ctx-3" {
		t.Errorf("expected first event to be ctx-3, got %s", all[0].ID)
	}

	limited, err := ListWorkContextEvents(db, 2)
	if err != nil {
		t.Fatalf("listing work context events with limit: %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("expected 2 events with limit, got %d", len(limited))
	}
}
