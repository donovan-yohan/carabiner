package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/spf13/cobra"
)

var (
	contextSetWorkItem string
	contextSetSpec     string
	contextShowJSON    bool
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage branch-scoped work context",
	Long:  "Commands for managing work context that is scoped to the current git branch.",
}

var contextSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set work context for current branch",
	Long:  "Set the work item reference and optional spec reference for the current branch.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if contextSetWorkItem == "" {
			return fmt.Errorf("work item ref is required")
		}

		ctx := carabiner.NewWorkContext(contextSetWorkItem, contextSetSpec)
		if err := carabiner.SetWorkContext(ctx); err != nil {
			return fmt.Errorf("setting work context: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Set work context for branch %s:\n", ctx.ContextBranch)
		fmt.Fprintf(os.Stdout, "  Work Item: %s\n", ctx.WorkItemRef)
		if ctx.SpecRef != "" {
			fmt.Fprintf(os.Stdout, "  Spec: %s\n", ctx.SpecRef)
		}
		fmt.Fprintf(os.Stdout, "  Set At: %s\n", ctx.SetAt)
		fmt.Fprintf(os.Stdout, "  Source: %s\n", ctx.Source)

		if db != nil {
			event := &events.WorkContextEvent{
				ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
				Timestamp:   time.Now(),
				WorkItemRef: ctx.WorkItemRef,
				SpecRef:     ctx.SpecRef,
				Branch:      ctx.ContextBranch,
				Source:      ctx.Source,
			}
			if err := events.AppendWorkContextEvent(db, event); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to record context event: %v\n", err)
			}
		}

		return nil
	},
}

var contextShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current work context",
	Long:  "Display the work context for the current branch.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := carabiner.GetWorkContext()
		if err != nil {
			return fmt.Errorf("getting work context: %w", err)
		}

		if ctx.WorkItemRef == "" {
			if contextShowJSON {
				return fmt.Errorf("no work context set")
			}
			fmt.Fprintln(os.Stdout, "No work context set for current branch.")
			return nil
		}

		if contextShowJSON {
			data, _ := json.Marshal(map[string]string{
				"WorkItemRef":   ctx.WorkItemRef,
				"SpecRef":       ctx.SpecRef,
				"ContextBranch": ctx.ContextBranch,
				"SetAt":         ctx.SetAt,
				"Source":        ctx.Source,
			})
			fmt.Fprintln(os.Stdout, string(data))
			return nil
		}

		fmt.Fprintf(os.Stdout, "Work Context for branch %s:\n", ctx.ContextBranch)
		fmt.Fprintf(os.Stdout, "  Work Item: %s\n", ctx.WorkItemRef)
		if ctx.SpecRef != "" {
			fmt.Fprintf(os.Stdout, "  Spec: %s\n", ctx.SpecRef)
		}
		fmt.Fprintf(os.Stdout, "  Set At: %s\n", ctx.SetAt)
		fmt.Fprintf(os.Stdout, "  Source: %s\n", ctx.Source)

		return nil
	},
}

var contextClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear work context for current branch",
	Long:  "Remove the work context from the current branch.",
	RunE: func(cmd *cobra.Command, args []string) error {
		result := carabiner.ClearWorkContext()

		if db != nil {
			event := &events.WorkContextEvent{
				ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
				Timestamp:   time.Now(),
				WorkItemRef: "",
				SpecRef:     "",
				Branch:      "",
				Source:      "clear",
			}
			if err := events.AppendWorkContextEvent(db, event); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to record clear event: %v\n", err)
			}
		}

		if !result.ClearSucceeded {
			fmt.Fprintf(os.Stderr, "warning: failed to unset some keys: %v\n", result.FailedKeys)
		}

		fmt.Fprintln(os.Stdout, "Cleared work context for current branch.")
		return nil
	},
}

func init() {
	contextSetCmd.Flags().StringVar(&contextSetWorkItem, "work-item", "", "work item reference (required)")
	contextSetCmd.Flags().StringVar(&contextSetSpec, "spec", "", "spec reference (optional)")

	contextShowCmd.Flags().BoolVar(&contextShowJSON, "json", false, "output in JSON format")

	contextCmd.AddCommand(contextSetCmd)
	contextCmd.AddCommand(contextShowCmd)
	contextCmd.AddCommand(contextClearCmd)
	rootCmd.AddCommand(contextCmd)
}
