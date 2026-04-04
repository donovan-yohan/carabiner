package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/donovan-yohan/carabiner/internal/carabiner/agentlytics"
	"github.com/donovan-yohan/carabiner/internal/carabiner/git"
	"github.com/spf13/cobra"
)

var doctorJSON bool

// DoctorReport is the structured output of carabiner doctor.
type DoctorReport struct {
	GitRepo     bool               `json:"git_repo"`
	GitAI       GitAIStatus        `json:"git_ai"`
	Agentlytics *agentlytics.Health `json:"agentlytics"`
	Ready       bool               `json:"ready"`
}

// GitAIStatus reports git-ai availability.
type GitAIStatus struct {
	NotesRefExists bool `json:"notes_ref_exists"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check data source availability",
	Long:  "Diagnose which data sources (git, git-ai, agentlytics) are available for forensic queries.",
	RunE:  runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	report := &DoctorReport{}

	// Check git
	report.GitRepo = git.IsInsideWorkTree()

	// Check git-ai
	if report.GitRepo {
		report.GitAI.NotesRefExists = git.HasNotesRef("ai")
	}

	// Check agentlytics
	agentlyticsPath := agentlytics.DefaultCachePath()
	report.Agentlytics = agentlytics.CheckHealth(agentlyticsPath)

	// Ready = all required sources present (agentlytics must also have valid schema)
	report.Ready = report.GitRepo && report.GitAI.NotesRefExists && report.Agentlytics.Found && report.Agentlytics.SchemaValid

	if doctorJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	return printDoctorText(report)
}

func printDoctorText(r *DoctorReport) error {
	fmt.Println("carabiner doctor")
	fmt.Println("================")
	fmt.Println()

	printCheck("Git repository", r.GitRepo, "", `Not inside a git repository.
Run this command from within a git repo.`)

	printCheck("git-ai notes (refs/notes/ai)", r.GitAI.NotesRefExists, "", `No git-ai notes found in this repo.
git-ai tracks which AI agent wrote each line of code.
Install: https://github.com/git-ai-project/git-ai
Then run: git-ai init`)

	agentlyticsOK := r.Agentlytics.Found && r.Agentlytics.SchemaValid
	detail := ""
	if agentlyticsOK {
		detail = fmt.Sprintf("%d sessions indexed", r.Agentlytics.ChatCount)
	}
	printCheck("agentlytics cache", agentlyticsOK, detail, fmt.Sprintf(`agentlytics cache not found at %s
agentlytics records AI session data from all agents.
Install: https://github.com/f/agentlytics
Then run: npx agentlytics`, r.Agentlytics.Path))

	fmt.Println()
	if r.Ready {
		fmt.Println("Ready. Run: carabiner why <file>:<line>")
	} else {
		fmt.Println("Not ready. Install missing data sources above.")
	}

	if !r.Ready {
		return fmt.Errorf("not ready: install missing data sources above")
	}
	return nil
}

func printCheck(name string, ok bool, detail string, fixMsg string) {
	if ok {
		marker := "[OK]"
		line := fmt.Sprintf("  %s %s", marker, name)
		if detail != "" {
			line += " (" + detail + ")"
		}
		fmt.Println(line)
	} else {
		fmt.Printf("  [--] %s\n", name)
		// Indent the fix message
		for _, line := range splitLines(fixMsg) {
			fmt.Printf("       %s\n", line)
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
