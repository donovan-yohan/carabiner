package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/stretchr/testify/require"
)

func TestContextSet_EmitsWorkContextEvent(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-context-test-*")

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	dummyFile := filepath.Join(repoDir, "README.md")
	require.NoError(os.WriteFile(dummyFile, []byte("# Test"), 0644))

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	initResult := runCLI(t, bin, repoDir, "init")
	require.Equal(0, initResult.exitCode, "stderr: %s", initResult.stderr)
	require.NoError(initResult.err)

	dbPath := filepath.Join(repoDir, ".carabiner", "carabiner.db")
	db, err := events.InitDB(dbPath)
	require.NoError(err)
	defer db.Close()

	setResult := runCLI(t, bin, repoDir, "context", "set", "--work-item", "TEST-123", "--spec", "spec-456")
	require.Equal(0, setResult.exitCode, "stderr: %s", setResult.stderr)
	require.NoError(setResult.err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM work_context_events").Scan(&count)
	require.NoError(err)
	require.Equal(1, count, "Expected 1 work_context_event")

	var event events.WorkContextEvent
	err = db.QueryRow("SELECT id, timestamp, work_item_ref, spec_ref, branch, source FROM work_context_events LIMIT 1").Scan(
		&event.ID,
		&event.Timestamp,
		&event.WorkItemRef,
		&event.SpecRef,
		&event.Branch,
		&event.Source,
	)
	require.NoError(err)
	require.Equal("TEST-123", event.WorkItemRef)
	require.Equal("spec-456", event.SpecRef)
	require.Equal("explicit", event.Source)
	require.NotEmpty(event.Branch)
	require.NotEmpty(event.ID)
	require.False(event.Timestamp.IsZero())
}

func TestContextClear_EmitsWorkContextEvent(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-context-test-*")

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	dummyFile := filepath.Join(repoDir, "README.md")
	require.NoError(os.WriteFile(dummyFile, []byte("# Test"), 0644))

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	initResult := runCLI(t, bin, repoDir, "init")
	require.Equal(0, initResult.exitCode, "stderr: %s", initResult.stderr)
	require.NoError(initResult.err)

	dbPath := filepath.Join(repoDir, ".carabiner", "carabiner.db")
	db, err := events.InitDB(dbPath)
	require.NoError(err)
	defer db.Close()

	setResult := runCLI(t, bin, repoDir, "context", "set", "--work-item", "TEST-789")
	require.Equal(0, setResult.exitCode, "stderr: %s", setResult.stderr)
	require.NoError(setResult.err)

	clearResult := runCLI(t, bin, repoDir, "context", "clear")
	require.Equal(0, clearResult.exitCode, "stderr: %s", clearResult.stderr)
	require.NoError(clearResult.err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM work_context_events").Scan(&count)
	require.NoError(err)
	require.Equal(2, count, "Expected 2 work_context_events (set + clear)")

	rows, err := db.Query("SELECT work_item_ref, spec_ref, branch, source FROM work_context_events ORDER BY timestamp")
	require.NoError(err)
	defer rows.Close()

	require.True(rows.Next())
	var workItemRef, specRef, branch, source string
	err = rows.Scan(&workItemRef, &specRef, &branch, &source)
	require.NoError(err)
	require.Equal("TEST-789", workItemRef)
	require.Equal("explicit", source)

	require.True(rows.Next())
	err = rows.Scan(&workItemRef, &specRef, &branch, &source)
	require.NoError(err)
	require.Empty(workItemRef)
	require.Empty(specRef)
	require.Empty(branch)
	require.Equal("clear", source)
}

func TestMetadataPayload_IncludesWorkContext(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-metadata-test-*")

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	dummyFile := filepath.Join(repoDir, "README.md")
	require.NoError(os.WriteFile(dummyFile, []byte("# Test"), 0644))

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	initResult := runCLI(t, bin, repoDir, "init")
	require.Equal(0, initResult.exitCode, "stderr: %s", initResult.stderr)
	require.NoError(initResult.err)

	cfgDir := filepath.Join(repoDir, ".carabiner")
	enforceYAML := `version: 1
tools:
  pass:
    enabled: true
    command: true
    args: []
behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: false
`
	require.NoError(os.WriteFile(filepath.Join(cfgDir, "enforce.yaml"), []byte(enforceYAML), 0644))

	dbPath := filepath.Join(repoDir, ".carabiner", "carabiner.db")
	db, err := events.InitDB(dbPath)
	require.NoError(err)
	defer db.Close()

	setResult := runCLI(t, bin, repoDir, "context", "set", "--work-item", "META-TEST", "--spec", "meta-spec")
	require.Equal(0, setResult.exitCode, "stderr: %s", setResult.stderr)
	require.NoError(setResult.err)

	enforceResult := runCLI(t, bin, repoDir, "enforce", "--all", "--config-dir", filepath.Join(repoDir, ".carabiner"))
	require.Equal(0, enforceResult.exitCode, "stderr: %s", enforceResult.stderr)
	require.NoError(enforceResult.err)

	var metadata string
	err = db.QueryRow("SELECT metadata FROM events WHERE command = 'enforce' ORDER BY timestamp DESC LIMIT 1").Scan(&metadata)
	require.NoError(err)

	var meta map[string]interface{}
	err = json.Unmarshal([]byte(metadata), &meta)
	require.NoError(err)

	workItemRef, ok := meta["workItemRef"]
	require.True(ok, "Expected workItemRef in metadata")
	require.Equal("META-TEST", workItemRef)

	specRef, ok := meta["specRef"]
	require.True(ok, "Expected specRef in metadata")
	require.Equal("meta-spec", specRef)

	contextBranch, ok := meta["contextBranch"]
	require.True(ok, "Expected contextBranch in metadata")
	require.NotEmpty(contextBranch)
}

func TestMetadataPayload_NoWorkContext(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-metadata-test-*")

	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	dummyFile := filepath.Join(repoDir, "README.md")
	require.NoError(os.WriteFile(dummyFile, []byte("# Test"), 0644))

	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = repoDir
	require.NoError(cmd.Run())

	initResult := runCLI(t, bin, repoDir, "init")
	require.Equal(0, initResult.exitCode, "stderr: %s", initResult.stderr)
	require.NoError(initResult.err)

	cfgDir := filepath.Join(repoDir, ".carabiner")
	enforceYAML := `version: 1
tools:
  pass:
    enabled: true
    command: true
    args: []
behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: false
`
	require.NoError(os.WriteFile(filepath.Join(cfgDir, "enforce.yaml"), []byte(enforceYAML), 0644))

	dbPath := filepath.Join(repoDir, ".carabiner", "carabiner.db")
	db, err := events.InitDB(dbPath)
	require.NoError(err)
	defer db.Close()

	enforceResult := runCLI(t, bin, repoDir, "enforce", "--all", "--config-dir", filepath.Join(repoDir, ".carabiner"))
	require.Equal(0, enforceResult.exitCode, "stderr: %s", enforceResult.stderr)
	require.NoError(enforceResult.err)

	var metadata string
	err = db.QueryRow("SELECT metadata FROM events WHERE command = 'enforce' ORDER BY timestamp DESC LIMIT 1").Scan(&metadata)
	require.NoError(err)

	var meta map[string]interface{}
	err = json.Unmarshal([]byte(metadata), &meta)
	require.NoError(err)

	_, hasWorkItem := meta["workItemRef"]
	_, hasSpec := meta["specRef"]
	_, hasBranch := meta["contextBranch"]

	require.False(hasWorkItem, "Should not have workItemRef when no context set")
	require.False(hasSpec, "Should not have specRef when no context set")
	require.False(hasBranch, "Should not have contextBranch when no context set")
}
