package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/agentlytics"
	"github.com/donovan-yohan/carabiner/internal/carabiner/dossier"
	"github.com/donovan-yohan/carabiner/internal/carabiner/git"
	"github.com/spf13/cobra"
)

var whyJSON bool
var whyRev string

var whyCmd = &cobra.Command{
	Use:   "why <file>:<line>",
	Short: "Forensic dossier for a line of code",
	Long:  "Trace a line of code back to the AI agent session that wrote it, including model, tool, and session metadata.",
	Args:  cobra.ExactArgs(1),
	RunE:  runWhy,
}

func init() {
	whyCmd.Flags().BoolVar(&whyJSON, "json", false, "output as JSON")
	whyCmd.Flags().StringVar(&whyRev, "rev", "", "git revision to blame against (default: working tree)")
	rootCmd.AddCommand(whyCmd)
}

func runWhy(cmd *cobra.Command, args []string) error {
	file, line, err := parseFileLine(args[0])
	if err != nil {
		return err
	}

	// Fail-fast: must be in a git repo
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not inside a git repository")
	}

	// Fail-fast: git-ai notes must exist
	if !git.HasNotesRef("ai") {
		fmt.Fprintln(os.Stderr, `No git-ai notes found in this repo.
git-ai tracks which AI agent wrote each line of code.
Install: https://github.com/git-ai-project/git-ai
Then run: git-ai init`)
		os.Exit(1)
	}

	// Find agentlytics cache (optional enrichment)
	agentlyticsPath := agentlytics.DefaultCachePath()
	if _, err := os.Stat(agentlyticsPath); err != nil {
		fmt.Fprintln(os.Stderr, `Warning: agentlytics cache not found at `+agentlyticsPath+`
agentlytics records AI session data from all agents.
Install: https://github.com/f/agentlytics
Then run: agentlytics scan
Continuing with git-ai data only.
`)
		agentlyticsPath = ""
	}

	builder := dossier.NewBuilder(agentlyticsPath)
	d, err := builder.Build(file, line, whyRev)
	if err != nil {
		return err
	}

	if whyJSON {
		return outputJSON(d)
	}
	return outputText(d)
}

func parseFileLine(arg string) (string, int, error) {
	lastColon := strings.LastIndex(arg, ":")
	if lastColon == -1 || lastColon == len(arg)-1 {
		return "", 0, fmt.Errorf("invalid format %q, expected <file>:<line>", arg)
	}

	file := arg[:lastColon]
	lineStr := arg[lastColon+1:]

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid line number %q: %w", lineStr, err)
	}
	if line < 1 {
		return "", 0, fmt.Errorf("line number must be >= 1, got %d", line)
	}

	return file, line, nil
}

func outputJSON(d *carabiner.Dossier) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

func outputText(d *carabiner.Dossier) error {
	fmt.Printf("LINE: %s:%d (introduced in commit %s)\n", d.File, d.Line, shortSHA(d.Blame.CommitSHA))
	fmt.Printf("AUTHOR: %s\n", d.Blame.Author)
	fmt.Printf("CONFIDENCE: %s\n", d.OverallConfidence)
	fmt.Println()

	if d.Session != nil {
		fmt.Printf("SESSION: %s session %s\n", d.Session.Tool, d.Session.ID)
		if d.Session.Model != "" {
			fmt.Printf("MODEL: %s\n", d.Session.Model)
		}
		if d.Session.Name != "" {
			fmt.Printf("NAME: %s\n", d.Session.Name)
		}
		if d.Session.Source != "" {
			fmt.Printf("SOURCE: %s\n", d.Session.Source)
		}
		if !d.Session.StartedAt.IsZero() {
			fmt.Printf("PERIOD: %s to %s\n",
				d.Session.StartedAt.Format("2006-01-02 15:04"),
				d.Session.EndedAt.Format("2006-01-02 15:04"))
		}
	} else {
		fmt.Println("SESSION: none (human-authored or pre-git-ai)")
	}

	fmt.Println()
	fmt.Println("ATTRIBUTION CHAIN:")
	for _, hop := range d.Hops {
		marker := "[OK]"
		if hop.Confidence == carabiner.ConfidenceMissing {
			marker = "[--]"
		}
		fmt.Printf("  %s %s: %s\n", marker, hop.Name, hop.Detail)
	}

	return nil
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
