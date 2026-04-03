package carabiner

import (
	"strings"
	"testing"
)

func TestRenderPreCommitHook(t *testing.T) {
	hook := RenderPreCommitHook()

	if !strings.HasPrefix(hook, "#!/bin/sh") {
		t.Error("pre-commit hook must start with shebang")
	}

	if !strings.Contains(hook, "carabiner context show") {
		t.Error("pre-commit hook must call carabiner context show")
	}

	if !strings.Contains(hook, "exit 1") {
		t.Error("pre-commit hook must exit with error code on failure")
	}
}

func TestRenderCommitMsgHook(t *testing.T) {
	hook := RenderCommitMsgHook()

	if !strings.HasPrefix(hook, "#!/bin/sh") {
		t.Error("commit-msg hook must start with shebang")
	}

	if !strings.Contains(hook, "carabiner context show --json") {
		t.Error("commit-msg hook must call carabiner context show --json")
	}

	if !strings.Contains(hook, "Carabiner-Work-Item:") {
		t.Error("commit-msg hook must append Carabiner-Work-Item trailer")
	}

	if !strings.Contains(hook, "Carabiner-Spec:") {
		t.Error("commit-msg hook must append Carabiner-Spec trailer")
	}

	if !strings.Contains(hook, "Carabiner-Context-Branch:") {
		t.Error("commit-msg hook must append Carabiner-Context-Branch trailer")
	}
}

func TestHookScriptsAreExecutable(t *testing.T) {
	preCommit := RenderPreCommitHook()
	commitMsg := RenderCommitMsgHook()

	if !strings.HasPrefix(preCommit, "#!/bin/sh") {
		t.Error("pre-commit hook must have shebang for executability")
	}

	if !strings.HasPrefix(commitMsg, "#!/bin/sh") {
		t.Error("commit-msg hook must have shebang for executability")
	}
}
