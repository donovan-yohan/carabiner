package telemetry

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	_ "modernc.org/sqlite"
)

func TestImportAgentlytics(t *testing.T) {
	// Create temp agentlytics source DB
	sourcePath := filepath.Join(t.TempDir(), "cache.db")
	sourceDB, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		t.Fatalf("opening source db: %v", err)
	}
	defer sourceDB.Close()

	// Create sessions table with test data
	_, err = sourceDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at TEXT,
			editor TEXT,
			repo_path TEXT,
			model TEXT,
			message_count INTEGER
		);

		CREATE TABLE messages (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			role TEXT,
			content TEXT
		);

		INSERT INTO sessions (id, created_at, editor, repo_path, model, message_count) VALUES
			('session-1', '2024-01-15T10:30:00Z', 'claude-code', '/home/user/project1', 'claude-3-opus', 5),
			('session-2', '2024-01-15T11:00:00Z', 'cursor', '/home/user/project2', 'gpt-4', 3),
			('session-3', '2024-01-15T11:30:00Z', 'zed', '/home/user/project3', 'claude-3-sonnet', 7);
	`)
	if err != nil {
		t.Fatalf("creating test data: %v", err)
	}

	// Create target workflow_events DB
	targetPath := filepath.Join(t.TempDir(), "carabiner.db")
	targetDB, err := events.InitDB(targetPath)
	if err != nil {
		t.Fatalf("initializing target db: %v", err)
	}
	defer targetDB.Close()

	// Import sessions
	imported, err := ImportAgentlytics(targetDB, AgentlyticsImportOptions{
		SourcePath: sourcePath,
	})
	if err != nil {
		t.Fatalf("importing agentlytics: %v", err)
	}

	if imported != 3 {
		t.Errorf("expected 3 imported sessions, got %d", imported)
	}

	// Verify workflow events were created
	workflowEvents, err := events.ListWorkflowEvents(targetDB, "", 0)
	if err != nil {
		t.Fatalf("listing workflow events: %v", err)
	}

	if len(workflowEvents) != 3 {
		t.Errorf("expected 3 workflow events, got %d", len(workflowEvents))
	}

	// Verify first session
	var found bool
	for _, event := range workflowEvents {
		if event.ExternalSessionID == "session-1" {
			found = true
			if event.ID != "agentlytics:session-1" {
				t.Errorf("expected ID 'agentlytics:session-1', got '%s'", event.ID)
			}
			if event.Workflow != "agentlytics" {
				t.Errorf("expected workflow 'agentlytics', got '%s'", event.Workflow)
			}
			if event.EventType != "session_imported" {
				t.Errorf("expected event_type 'session_imported', got '%s'", event.EventType)
			}
			if event.Agent != "claude-code" {
				t.Errorf("expected agent 'claude-code', got '%s'", event.Agent)
			}
			if event.Model != "claude-3-opus" {
				t.Errorf("expected model 'claude-3-opus', got '%s'", event.Model)
			}
			if event.RepoPath != "/home/user/project1" {
				t.Errorf("expected repo_path '/home/user/project1', got '%s'", event.RepoPath)
			}
			break
		}
	}

	if !found {
		t.Error("session-1 not found in workflow events")
	}
}

func TestImportAgentlyticsWatermark(t *testing.T) {
	// Create temp agentlytics source DB
	sourcePath := filepath.Join(t.TempDir(), "cache.db")
	sourceDB, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		t.Fatalf("opening source db: %v", err)
	}
	defer sourceDB.Close()

	_, err = sourceDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at TEXT,
			editor TEXT,
			repo_path TEXT,
			model TEXT,
			message_count INTEGER
		);

		INSERT INTO sessions (id, created_at, editor, repo_path, model, message_count) VALUES
			('session-1', '2024-01-15T10:30:00Z', 'claude-code', '/home/user/project1', 'claude-3-opus', 5);
	`)
	if err != nil {
		t.Fatalf("creating test data: %v", err)
	}

	// Create target workflow_events DB
	targetPath := filepath.Join(t.TempDir(), "carabiner.db")
	targetDB, err := events.InitDB(targetPath)
	if err != nil {
		t.Fatalf("initializing target db: %v", err)
	}
	defer targetDB.Close()

	// First import
	imported, err := ImportAgentlytics(targetDB, AgentlyticsImportOptions{
		SourcePath: sourcePath,
	})
	if err != nil {
		t.Fatalf("first import: %v", err)
	}

	if imported != 1 {
		t.Errorf("expected 1 imported session on first import, got %d", imported)
	}

	// Second import (should skip due to watermark)
	imported, err = ImportAgentlytics(targetDB, AgentlyticsImportOptions{
		SourcePath: sourcePath,
	})
	if err != nil {
		t.Fatalf("second import: %v", err)
	}

	if imported != 0 {
		t.Errorf("expected 0 imported sessions on second import (watermark), got %d", imported)
	}

	// Verify only one event exists
	workflowEvents, err := events.ListWorkflowEvents(targetDB, "", 0)
	if err != nil {
		t.Fatalf("listing workflow events: %v", err)
	}

	if len(workflowEvents) != 1 {
		t.Errorf("expected 1 workflow event after duplicate import, got %d", len(workflowEvents))
	}
}

func TestImportAgentlyticsWithLimit(t *testing.T) {
	// Create temp agentlytics source DB
	sourcePath := filepath.Join(t.TempDir(), "cache.db")
	sourceDB, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		t.Fatalf("opening source db: %v", err)
	}
	defer sourceDB.Close()

	_, err = sourceDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at TEXT,
			editor TEXT,
			repo_path TEXT,
			model TEXT,
			message_count INTEGER
		);

		INSERT INTO sessions (id, created_at, editor, repo_path, model, message_count) VALUES
			('session-1', '2024-01-15T10:30:00Z', 'claude-code', '/home/user/project1', 'claude-3-opus', 5),
			('session-2', '2024-01-15T11:00:00Z', 'cursor', '/home/user/project2', 'gpt-4', 3),
			('session-3', '2024-01-15T11:30:00Z', 'zed', '/home/user/project3', 'claude-3-sonnet', 7);
	`)
	if err != nil {
		t.Fatalf("creating test data: %v", err)
	}

	// Create target workflow_events DB
	targetPath := filepath.Join(t.TempDir(), "carabiner.db")
	targetDB, err := events.InitDB(targetPath)
	if err != nil {
		t.Fatalf("initializing target db: %v", err)
	}
	defer targetDB.Close()

	// Import with limit
	imported, err := ImportAgentlytics(targetDB, AgentlyticsImportOptions{
		SourcePath: sourcePath,
		Limit:      2,
	})
	if err != nil {
		t.Fatalf("importing with limit: %v", err)
	}

	if imported != 2 {
		t.Errorf("expected 2 imported sessions with limit, got %d", imported)
	}
}

func TestImportAgentlyticsSchemaDrift(t *testing.T) {
	// Create temp agentlytics source DB with extra columns
	sourcePath := filepath.Join(t.TempDir(), "cache.db")
	sourceDB, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		t.Fatalf("opening source db: %v", err)
	}
	defer sourceDB.Close()

	_, err = sourceDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at TEXT,
			editor TEXT,
			repo_path TEXT,
			model TEXT,
			message_count INTEGER,
			extra_field_1 TEXT,
			extra_field_2 INTEGER,
			another_field REAL
		);

		INSERT INTO sessions (id, created_at, editor, repo_path, model, message_count, extra_field_1, extra_field_2, another_field) VALUES
			('session-1', '2024-01-15T10:30:00Z', 'claude-code', '/home/user/project1', 'claude-3-opus', 5, 'extra-value', 42, 3.14);
	`)
	if err != nil {
		t.Fatalf("creating test data: %v", err)
	}

	// Create target workflow_events DB
	targetPath := filepath.Join(t.TempDir(), "carabiner.db")
	targetDB, err := events.InitDB(targetPath)
	if err != nil {
		t.Fatalf("initializing target db: %v", err)
	}
	defer targetDB.Close()

	// Import should succeed despite unknown columns
	imported, err := ImportAgentlytics(targetDB, AgentlyticsImportOptions{
		SourcePath: sourcePath,
	})
	if err != nil {
		t.Fatalf("importing with schema drift: %v", err)
	}

	if imported != 1 {
		t.Errorf("expected 1 imported session, got %d", imported)
	}

	// Verify workflow event was created
	workflowEvents, err := events.ListWorkflowEvents(targetDB, "", 0)
	if err != nil {
		t.Fatalf("listing workflow events: %v", err)
	}

	if len(workflowEvents) != 1 {
		t.Errorf("expected 1 workflow event, got %d", len(workflowEvents))
	}
}

func TestImportAgentlyticsMissingTable(t *testing.T) {
	// Create temp agentlytics source DB without sessions table
	sourcePath := filepath.Join(t.TempDir(), "cache.db")
	sourceDB, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		t.Fatalf("opening source db: %v", err)
	}
	defer sourceDB.Close()

	_, err = sourceDB.Exec(`
		CREATE TABLE other_table (
			id TEXT PRIMARY KEY
		);
	`)
	if err != nil {
		t.Fatalf("creating test data: %v", err)
	}

	// Create target workflow_events DB
	targetPath := filepath.Join(t.TempDir(), "carabiner.db")
	targetDB, err := events.InitDB(targetPath)
	if err != nil {
		t.Fatalf("initializing target db: %v", err)
	}
	defer targetDB.Close()

	// Import should fail gracefully
	_, err = ImportAgentlytics(targetDB, AgentlyticsImportOptions{
		SourcePath: sourcePath,
	})
	if err == nil {
		t.Error("expected error when sessions table is missing")
	}
}

func TestDefaultAgentlyticsCachePath(t *testing.T) {
	path := DefaultAgentlyticsCachePath()
	if path == "" {
		t.Error("expected non-empty default path")
	}

	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got '%s'", path)
	}

	if filepath.Base(filepath.Dir(path)) != ".agentlytics" {
		t.Errorf("expected path to be in .agentlytics directory, got '%s'", path)
	}

	if filepath.Base(path) != "cache.db" {
		t.Errorf("expected filename 'cache.db', got '%s'", filepath.Base(path))
	}
}

func TestImportAgentlyticsAlternateTimestampFormat(t *testing.T) {
	// Create temp agentlytics source DB with alternate timestamp format
	sourcePath := filepath.Join(t.TempDir(), "cache.db")
	sourceDB, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		t.Fatalf("opening source db: %v", err)
	}
	defer sourceDB.Close()

	_, err = sourceDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			created_at TEXT,
			editor TEXT,
			repo_path TEXT,
			model TEXT,
			message_count INTEGER
		);

		INSERT INTO sessions (id, created_at, editor, repo_path, model, message_count) VALUES
			('session-1', '2024-01-15 10:30:00', 'claude-code', '/home/user/project1', 'claude-3-opus', 5);
	`)
	if err != nil {
		t.Fatalf("creating test data: %v", err)
	}

	// Create target workflow_events DB
	targetPath := filepath.Join(t.TempDir(), "carabiner.db")
	targetDB, err := events.InitDB(targetPath)
	if err != nil {
		t.Fatalf("initializing target db: %v", err)
	}
	defer targetDB.Close()

	// Import should handle alternate timestamp format
	imported, err := ImportAgentlytics(targetDB, AgentlyticsImportOptions{
		SourcePath: sourcePath,
	})
	if err != nil {
		t.Fatalf("importing with alternate timestamp: %v", err)
	}

	if imported != 1 {
		t.Errorf("expected 1 imported session, got %d", imported)
	}

	// Verify timestamp was parsed
	workflowEvents, err := events.ListWorkflowEvents(targetDB, "", 0)
	if err != nil {
		t.Fatalf("listing workflow events: %v", err)
	}

	if len(workflowEvents) != 1 {
		t.Errorf("expected 1 workflow event, got %d", len(workflowEvents))
		return
	}

	expectedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !workflowEvents[0].Timestamp.Equal(expectedTime) {
		t.Errorf("expected timestamp %v, got %v", expectedTime, workflowEvents[0].Timestamp)
	}
}
