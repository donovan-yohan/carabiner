package enforce

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ToolResult represents the result of running a single enforcement tool.
type ToolResult struct {
	Name     string
	ExitCode int
	Output   string
	Duration time.Duration
	Error    error
}

// EnforceResult represents the aggregated result of running all enforcement tools.
type EnforceResult struct {
	Results  []ToolResult
	ExitCode int
}

// RunTool executes a single tool and returns the result.
func RunTool(ctx context.Context, name string, cfg ToolConfig) (*ToolResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &ToolResult{
		Name:     name,
		Duration: duration,
		Output:   string(output),
	}

	if err != nil {
		if ctx.Err() != nil {
			result.Error = fmt.Errorf("tool execution timed out or was cancelled: %w", ctx.Err())
			result.ExitCode = -1
			return result, result.Error
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}

		if execErr, ok := err.(*exec.Error); ok {
			if execErr.Err == exec.ErrNotFound {
				result.Error = fmt.Errorf("tool binary not found: %s", cfg.Command)
				result.ExitCode = 127
				return result, result.Error
			}
		}

		result.Error = fmt.Errorf("unexpected error running tool: %w", err)
		result.ExitCode = -1
		return result, result.Error
	}

	result.ExitCode = 0
	return result, nil
}

// RunEnforce runs all enabled tools sequentially and returns the aggregated result.
func RunEnforce(ctx context.Context, cfg *EnforceConfig, selectedTool string) (*EnforceResult, error) {
	result := &EnforceResult{
		Results:  []ToolResult{},
		ExitCode: 0,
	}

	var toolsToRun []string
	if selectedTool != "" {
		if _, exists := cfg.Tools[selectedTool]; !exists {
			result.ExitCode = 2
			return result, fmt.Errorf("tool not found in config: %s", selectedTool)
		}
		toolsToRun = []string{selectedTool}
	} else {
		for name, toolCfg := range cfg.Tools {
			if toolCfg.Enabled {
				toolsToRun = append(toolsToRun, name)
			}
		}
	}

	for _, name := range toolsToRun {
		toolCfg := cfg.Tools[name]

		toolCtx := ctx
		var cancel context.CancelFunc
		if toolCfg.Timeout > 0 {
			toolCtx, cancel = context.WithTimeout(ctx, time.Duration(toolCfg.Timeout)*time.Second)
			defer cancel()
		}

		toolResult, err := RunTool(toolCtx, name, toolCfg)
		if err != nil {
			result.Results = append(result.Results, *toolResult)
			result.ExitCode = 2
			return result, err
		}

		result.Results = append(result.Results, *toolResult)

		if toolResult.ExitCode != 0 {
			result.ExitCode = 1
		}

		if cfg.Behavior.StopOnFirstFail && toolResult.ExitCode != 0 {
			break
		}
	}

	return result, nil
}
