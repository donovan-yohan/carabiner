package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/enforce"
	"github.com/spf13/cobra"
)

var (
	enforceAll    bool
	enforceTool   string
	enforceOutput string
)

type enforceToolOutput struct {
	Name     string `json:"name"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output,omitempty"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

type enforceCommandOutput struct {
	Results  []enforceToolOutput `json:"results"`
	ExitCode int                 `json:"exit_code"`
}

var enforceCmd = &cobra.Command{
	Use:   "enforce [--all] [--tool name] [--output json|text]",
	Short: "Run configured enforcement checks",
	Long:  "Run configured enforcement tools from enforce.yaml and return pass/fail exit codes for automation.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if enforceAll && enforceTool != "" {
			fmt.Fprintf(os.Stderr, "--all and --tool cannot be used together\n")
			os.Exit(2)
		}

		if enforceOutput != "text" && enforceOutput != "json" {
			fmt.Fprintf(os.Stderr, "invalid --output value %q (allowed: text, json)\n", enforceOutput)
			os.Exit(2)
		}

		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}

		cfg, err := enforce.LoadEnforceConfig(cfgDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}

		if err := enforce.ValidateEnforceConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}

		selectedTool := enforceTool
		if enforceAll {
			selectedTool = ""
		}

		result, err := enforce.RunEnforce(context.Background(), cfg, selectedTool)
		if err != nil {
			if result != nil {
				printEnforceResult(result, enforceOutput)
			}
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}

		printEnforceResult(result, enforceOutput)

		switch result.ExitCode {
		case 0:
			if enforceOutput == "text" {
				fmt.Println("All checks passed")
			}
			return nil
		case 1:
			os.Exit(1)
		default:
			os.Exit(2)
		}

		return nil
	},
}

func printEnforceResult(result *enforce.EnforceResult, output string) {
	if result == nil {
		return
	}

	if output == "json" {
		jsonResult := enforceCommandOutput{
			Results:  make([]enforceToolOutput, 0, len(result.Results)),
			ExitCode: result.ExitCode,
		}

		for _, toolResult := range result.Results {
			entry := enforceToolOutput{
				Name:     toolResult.Name,
				ExitCode: toolResult.ExitCode,
				Output:   strings.TrimSpace(toolResult.Output),
				Duration: toolResult.Duration.String(),
			}
			if toolResult.Error != nil {
				entry.Error = toolResult.Error.Error()
			}
			jsonResult.Results = append(jsonResult.Results, entry)
		}

		payload, err := json.MarshalIndent(jsonResult, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to format JSON output: %v\n", err)
			return
		}
		fmt.Println(string(payload))
		return
	}

	for _, toolResult := range result.Results {
		status := "PASS"
		if toolResult.ExitCode != 0 {
			status = "FAIL"
		}

		fmt.Printf("[%s] %s (exit=%d, duration=%s)\n", status, toolResult.Name, toolResult.ExitCode, toolResult.Duration)
		if toolResult.Output != "" {
			fmt.Println(strings.TrimSpace(toolResult.Output))
		}
		if toolResult.Error != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", toolResult.Error)
		}
	}
}

func init() {
	enforceCmd.Flags().BoolVar(&enforceAll, "all", false, "run all enabled enforcement tools")
	enforceCmd.Flags().StringVar(&enforceTool, "tool", "", "run a specific enforcement tool")
	enforceCmd.Flags().StringVar(&enforceOutput, "output", "text", "output format: text or json")
	rootCmd.AddCommand(enforceCmd)
}
