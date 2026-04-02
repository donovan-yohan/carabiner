package enforce

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunTool_Pass(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := ToolConfig{
		Enabled: true,
		Command: "true",
		Args:    []string{},
	}

	result, err := RunTool(ctx, "true", cfg)
	assert.NoError(err)
	assert.Equal("true", result.Name)
	assert.Equal(0, result.ExitCode)
	assert.Equal("", result.Output)
	assert.NoError(result.Error)
	assert.Greater(result.Duration, time.Duration(0))
}

func TestRunTool_Fail(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := ToolConfig{
		Enabled: true,
		Command: "false",
		Args:    []string{},
	}

	result, err := RunTool(ctx, "false", cfg)
	assert.NoError(err)
	assert.Equal("false", result.Name)
	assert.Equal(1, result.ExitCode)
	assert.NoError(result.Error)
}

func TestRunTool_Timeout(t *testing.T) {
	assert := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cfg := ToolConfig{
		Enabled: true,
		Command: "sleep",
		Args:    []string{"10"},
	}

	result, err := RunTool(ctx, "sleep", cfg)
	assert.Error(err)
	assert.Equal("sleep", result.Name)
	assert.Error(result.Error)
	assert.Contains(result.Error.Error(), "timed out")
}

func TestRunTool_NotFound(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := ToolConfig{
		Enabled: true,
		Command: "nonexistent-binary-xyz123",
		Args:    []string{},
	}

	result, err := RunTool(ctx, "nonexistent", cfg)
	assert.Error(err)
	assert.Equal("nonexistent", result.Name)
	assert.Error(result.Error)
	assert.Contains(result.Error.Error(), "binary not found")
}

func TestRunEnforce_AllPass(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"true": {
				Enabled: true,
				Command: "true",
				Args:    []string{},
			},
			"echo": {
				Enabled: true,
				Command: "echo",
				Args:    []string{"hello"},
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   false,
			StopOnFirstFail: false,
			Parallel:        false,
		},
	}

	result, err := RunEnforce(ctx, cfg, "")
	assert.NoError(err)
	assert.Equal(0, result.ExitCode)
	assert.Len(result.Results, 2)

	for _, r := range result.Results {
		assert.Equal(0, r.ExitCode)
		assert.NoError(r.Error)
	}
}

func TestRunEnforce_OneFails(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"true": {
				Enabled: true,
				Command: "true",
				Args:    []string{},
			},
			"false": {
				Enabled: true,
				Command: "false",
				Args:    []string{},
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   false,
			StopOnFirstFail: false,
			Parallel:        false,
		},
	}

	result, err := RunEnforce(ctx, cfg, "")
	assert.NoError(err)
	assert.Equal(1, result.ExitCode)
	assert.Len(result.Results, 2)

	var passCount, failCount int
	for _, r := range result.Results {
		if r.ExitCode == 0 {
			passCount++
		} else {
			failCount++
		}
	}
	assert.Equal(1, passCount)
	assert.Equal(1, failCount)
}

func TestRunEnforce_ConfigError(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"missing": {
				Enabled: true,
				Command: "nonexistent-binary-xyz123",
				Args:    []string{},
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   false,
			StopOnFirstFail: false,
			Parallel:        false,
		},
	}

	result, err := RunEnforce(ctx, cfg, "")
	assert.Error(err)
	assert.Equal(2, result.ExitCode)
	assert.Len(result.Results, 1)
	assert.Error(result.Results[0].Error)
}

func TestRunEnforce_SingleTool(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"true": {
				Enabled: true,
				Command: "true",
				Args:    []string{},
			},
			"false": {
				Enabled: true,
				Command: "false",
				Args:    []string{},
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   false,
			StopOnFirstFail: false,
			Parallel:        false,
		},
	}

	result, err := RunEnforce(ctx, cfg, "true")
	assert.NoError(err)
	assert.Equal(0, result.ExitCode)
	assert.Len(result.Results, 1)
	assert.Equal("true", result.Results[0].Name)
	assert.Equal(0, result.Results[0].ExitCode)
}

func TestRunEnforce_SingleToolNotFound(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"true": {
				Enabled: true,
				Command: "true",
				Args:    []string{},
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   false,
			StopOnFirstFail: false,
			Parallel:        false,
		},
	}

	result, err := RunEnforce(ctx, cfg, "nonexistent-tool")
	assert.Error(err)
	assert.Equal(2, result.ExitCode)
	assert.Contains(err.Error(), "tool not found in config")
}

func TestRunEnforce_StopOnFirstFail(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"false": {
				Enabled: true,
				Command: "false",
				Args:    []string{},
			},
			"true": {
				Enabled: true,
				Command: "true",
				Args:    []string{},
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   false,
			StopOnFirstFail: true,
			Parallel:        false,
		},
	}

	result, err := RunEnforce(ctx, cfg, "")
	assert.NoError(err)
	assert.Equal(1, result.ExitCode)
	assert.Len(result.Results, 1)
	assert.Equal("false", result.Results[0].Name)
}

func TestRunEnforce_WithTimeout(t *testing.T) {
	assert := require.New(t)
	ctx := context.Background()

	cfg := &EnforceConfig{
		Version: 1,
		Tools: map[string]ToolConfig{
			"echo": {
				Enabled: true,
				Command: "echo",
				Args:    []string{"test"},
				Timeout: 5,
			},
		},
		Behavior: BehaviorConfig{
			FailOnWarning:   false,
			StopOnFirstFail: false,
			Parallel:        false,
		},
	}

	result, err := RunEnforce(ctx, cfg, "")
	assert.NoError(err)
	assert.Equal(0, result.ExitCode)
	assert.Len(result.Results, 1)
	assert.Contains(result.Results[0].Output, "test")
}
