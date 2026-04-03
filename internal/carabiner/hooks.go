package carabiner

// RenderPreCommitHook returns a shell script for the pre-commit hook
// that validates work context before allowing commits.
func RenderPreCommitHook() string {
	return `#!/bin/sh
carabiner context show >/dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "carabiner: no valid work context for current branch"
  echo "Run: carabiner context set --work-item <ref> [--spec <ref>]"
  exit 1
fi
`
}

// RenderCommitMsgHook returns a shell script for the commit-msg hook
// that appends carabiner trailers to commit messages.
func RenderCommitMsgHook() string {
	return `#!/bin/sh
# Read the commit message file
COMMIT_MSG_FILE=$1

# Get current context in JSON format
CONTEXT=$(carabiner context show --json 2>/dev/null)

if [ $? -ne 0 ]; then
  echo "carabiner: no valid work context for current branch"
  echo "Run: carabiner context set --work-item <ref> [--spec <ref>]"
  exit 1
fi

# Extract values from JSON (requires jq or basic parsing)
WORK_ITEM=$(echo "$CONTEXT" | grep -o '"WorkItemRef":"[^"]*"' | cut -d'"' -f4)
SPEC_REF=$(echo "$CONTEXT" | grep -o '"SpecRef":"[^"]*"' | cut -d'"' -f4)
CONTEXT_BRANCH=$(echo "$CONTEXT" | grep -o '"ContextBranch":"[^"]*"' | cut -d'"' -f4)

# Append trailers to commit message
echo "" >> "$COMMIT_MSG_FILE"
echo "Carabiner-Work-Item: $WORK_ITEM" >> "$COMMIT_MSG_FILE"

if [ -n "$SPEC_REF" ] && [ "$SPEC_REF" != "" ]; then
  echo "Carabiner-Spec: $SPEC_REF" >> "$COMMIT_MSG_FILE"
fi

echo "Carabiner-Context-Branch: $CONTEXT_BRANCH" >> "$COMMIT_MSG_FILE"
`
}
