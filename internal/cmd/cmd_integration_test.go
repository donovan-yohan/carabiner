package cmd

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
	err      error
}

func TestEnforce_AllPass(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-it-repo-*")
	cfgDir := filepath.Join(repoDir, ".carabiner")
	require.NoError(os.MkdirAll(cfgDir, 0o755))

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
	require.NoError(os.WriteFile(filepath.Join(cfgDir, "enforce.yaml"), []byte(enforceYAML), 0o644))

	result := runCLI(t, bin, repoDir, "enforce", "--all", "--config-dir", cfgDir)
	require.Equal(0, result.exitCode, "stderr: %s", result.stderr)
	require.NoError(result.err)
	require.Contains(result.stdout, "[PASS] pass")
	require.Contains(result.stdout, "All checks passed")
}

func TestEnforce_ToolFails(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-it-repo-*")
	cfgDir := filepath.Join(repoDir, ".carabiner")
	require.NoError(os.MkdirAll(cfgDir, 0o755))

	enforceYAML := `version: 1
tools:
  fail:
    enabled: true
    command: false
    args: []
behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: false
`
	require.NoError(os.WriteFile(filepath.Join(cfgDir, "enforce.yaml"), []byte(enforceYAML), 0o644))

	result := runCLI(t, bin, repoDir, "enforce", "--all", "--config-dir", cfgDir)
	require.Equal(1, result.exitCode)
	require.Error(result.err)
	require.Contains(result.stdout, "[FAIL] fail")
}

func TestEnforce_ToolFlagRunsSpecificTool(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-it-repo-*")
	cfgDir := filepath.Join(repoDir, ".carabiner")
	require.NoError(os.MkdirAll(cfgDir, 0o755))

	enforceYAML := `version: 1
tools:
  pass:
    enabled: true
    command: true
    args: []
  fail:
    enabled: true
    command: false
    args: []
behavior:
  fail_on_warning: true
  stop_on_first_failure: false
  parallel: false
`
	require.NoError(os.WriteFile(filepath.Join(cfgDir, "enforce.yaml"), []byte(enforceYAML), 0o644))

	result := runCLI(t, bin, repoDir, "enforce", "--tool", "pass", "--config-dir", cfgDir)
	require.Equal(0, result.exitCode, "stderr: %s", result.stderr)
	require.NoError(result.err)
	require.Contains(result.stdout, "[PASS] pass")
	require.NotContains(result.stdout, "fail")
}

func TestEvents_List(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-it-repo-*")

	initResult := runCLI(t, bin, repoDir, "init")
	require.Equal(0, initResult.exitCode, "stderr: %s", initResult.stderr)
	require.NoError(initResult.err)

	failedRecord := runCLI(t, bin, repoDir, "quality", "record", "--gate-id", "it-test", "--skip-extraction")
	require.Equal(2, failedRecord.exitCode)
	require.Error(failedRecord.err)
	require.Contains(failedRecord.stderr, "no input provided")

	var listResult cliResult
	for range 20 {
		listResult = runCLI(t, bin, repoDir, "events", "list", "--limit", "50")
		require.Equal(0, listResult.exitCode, "stderr: %s", listResult.stderr)
		if strings.Contains(listResult.stdout, "quality record") {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	require.Contains(listResult.stdout, "COMMAND")
	require.Contains(listResult.stdout, "quality record")
}

func TestInit_CreatesDB(t *testing.T) {
	require := require.New(t)

	bin := buildCarabinerBinary(t)
	repoDir := mustMkdirTemp(t, "carabiner-it-repo-*")

	result := runCLI(t, bin, repoDir, "init")
	require.Equal(0, result.exitCode, "stderr: %s", result.stderr)
	require.NoError(result.err)

	cfgDir := filepath.Join(repoDir, ".carabiner")
	require.DirExists(filepath.Join(cfgDir, "quality", "learnings"))
	require.DirExists(filepath.Join(cfgDir, "quality", "signals"))
	require.FileExists(filepath.Join(cfgDir, "config.yaml"))
	require.FileExists(filepath.Join(cfgDir, "carabiner.db"))
}

func buildCarabinerBinary(t *testing.T) string {
	t.Helper()
	require := require.New(t)

	root := projectRoot(t)
	buildDir := mustMkdirTemp(t, "carabiner-it-build-*")
	binPath := filepath.Join(buildDir, "carabiner")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/carabiner")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	require.NoError(err, "go build failed:\n%s", string(output))

	return binPath
}

func runCLI(t *testing.T, binaryPath, workDir string, args ...string) cliResult {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := cliResult{stdout: stdout.String(), stderr: stderr.String(), err: err}
	if err == nil {
		res.exitCode = 0
		return res
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.exitCode = exitErr.ExitCode()
		return res
	}

	t.Fatalf("failed to run %q with args %v: %v", binaryPath, args, err)
	return cliResult{}
}

func projectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	root := filepath.Clean(filepath.Join(wd, "../.."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("failed to resolve project root from %q: %v", wd, err)
	}
	return root
}

func mustMkdirTemp(t *testing.T, pattern string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", pattern)
	require.NoError(t, err)
	t.Cleanup(func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			t.Fatalf("cleanup temp dir %q: %v", dir, removeErr)
		}
	})
	return dir
}
