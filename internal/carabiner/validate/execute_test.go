package validate

import (
	"strings"
	"testing"

	"github.com/donovan-yohan/carabiner/internal/carabiner/enforce"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteValidations_ValidScript(t *testing.T) {
	validations := []enforce.ValidationConfig{
		{
			Name:   "test-validation",
			Script: "echo 'hello world'",
		},
	}

	results, runID, err := ExecuteValidations(validations)

	require.NoError(t, err)
	require.NotEmpty(t, runID, "runID should be generated")
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "test-validation", result.Name)
	assert.Equal(t, "echo 'hello world'", result.Script)
	assert.Equal(t, runID, result.RunID, "RunID should be consistent within call")
	assert.NoError(t, result.Error)
	assert.Contains(t, result.Question, "hello world")
	assert.NotZero(t, result.CreatedAt)
}

func TestExecuteValidations_InvalidScript(t *testing.T) {
	validations := []enforce.ValidationConfig{
		{
			Name:   "failing-validation",
			Script: "exit 1",
		},
	}

	results, runID, err := ExecuteValidations(validations)

	require.NoError(t, err)
	require.NotEmpty(t, runID)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "failing-validation", result.Name)
	assert.Error(t, result.Error, "Expected error for failing script")
	assert.Contains(t, result.Error.Error(), "script error")
}

func TestExecuteValidations_MultipleValidations(t *testing.T) {
	validations := []enforce.ValidationConfig{
		{
			Name:   "first",
			Script: "echo 'first output'",
		},
		{
			Name:   "second",
			Script: "echo 'second output'",
		},
		{
			Name:   "failing",
			Script: "exit 42",
		},
	}

	results, runID, err := ExecuteValidations(validations)

	require.NoError(t, err)
	require.NotEmpty(t, runID)
	require.Len(t, results, 3)

	// All results should have the same runID
	for _, result := range results {
		assert.Equal(t, runID, result.RunID, "All results should share the same runID")
	}

	// Check first result
	assert.Equal(t, "first", results[0].Name)
	assert.NoError(t, results[0].Error)
	assert.Contains(t, results[0].Question, "first output")

	// Check second result
	assert.Equal(t, "second", results[1].Name)
	assert.NoError(t, results[1].Error)
	assert.Contains(t, results[1].Question, "second output")

	// Check failing result
	assert.Equal(t, "failing", results[2].Name)
	assert.Error(t, results[2].Error)
}

func TestExecuteValidations_EmptyValidations(t *testing.T) {
	validations := []enforce.ValidationConfig{}

	results, runID, err := ExecuteValidations(validations)

	require.NoError(t, err)
	require.NotEmpty(t, runID)
	assert.Empty(t, results)
}

func TestExecuteValidations_RunIDConsistency(t *testing.T) {
	// Run multiple times and verify runIDs are unique
	validations := []enforce.ValidationConfig{
		{Name: "test", Script: "echo test"},
	}

	runIDs := make(map[string]bool)
	for i := 0; i < 5; i++ {
		_, runID, err := ExecuteValidations(validations)
		require.NoError(t, err)
		require.NotEmpty(t, runID)

		// Each run should have a unique runID
		assert.False(t, runIDs[runID], "runID should be unique across calls")
		runIDs[runID] = true
	}
}

func TestExecuteValidations_ScriptWithStderr(t *testing.T) {
	validations := []enforce.ValidationConfig{
		{
			Name:   "stderr-test",
			Script: "echo 'error message' >&2",
		},
	}

	results, _, err := ExecuteValidations(validations)

	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.NoError(t, result.Error, "Script with stderr but exit 0 should not be an error")
	assert.Contains(t, result.Question, "error message")
}

func TestExecuteValidations_ScriptWithArguments(t *testing.T) {
	validations := []enforce.ValidationConfig{
		{
			Name:   "args-test",
			Script: "echo $1 $2",
		},
	}

	results, _, err := ExecuteValidations(validations)

	require.NoError(t, err)
	require.Len(t, results, 1)

	// Script uses $1 and $2 which will be empty, so output should be empty or whitespace
	assert.True(t, strings.TrimSpace(results[0].Question) == "")
}
