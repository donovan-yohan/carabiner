package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/events"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var configDir string
var startTime time.Time
var db *sql.DB
var executedCommand string
var exitCode int

var rootCmd = &cobra.Command{
	Use:   "carabiner",
	Short: "Forensic query layer for AI-coded repos",
	Long:  "Carabiner joins git-ai attribution and agentlytics session data into a single forensic query: carabiner why <file>:<line>.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		startTime = time.Now()
		exitCode = 0
		executedCommand = commandLabel(cmd)

		if db != nil {
			return nil
		}

		cfgDir, err := carabiner.FindConfigDir(configDir)
		if err != nil {
			return nil
		}

		eventsDB, err := events.InitDB(filepath.Join(cfgDir, "carabiner.db"))
		if err != nil {
			return nil
		}

		db = eventsDB
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logEvent(cmd)
	},
}

func Execute() {
	executed, err := rootCmd.ExecuteC()
	if err != nil {
		exitCode = 2
		if executed != nil {
			executedCommand = commandLabel(executed)
		}
		logEvent(executed)
		fmt.Fprintln(os.Stderr, err)
		if db != nil {
			_ = db.Close()
		}
		os.Exit(2)
	}

	if db != nil {
		_ = db.Close()
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "path to .carabiner directory (default: walk up from CWD)")
}

func logEvent(cmd *cobra.Command) {
	if db == nil {
		return
	}

	event := &events.Event{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Command:    commandLabel(cmd),
		Args:       argsPayload(cmd),
		ExitCode:   exitCode,
		DurationMs: time.Since(startTime).Milliseconds(),
		RunID:      os.Getenv("CARABINER_RUN_ID"),
		Branch:     gitValue("rev-parse", "--abbrev-ref", "HEAD"),
		Commit:     gitValue("rev-parse", "HEAD"),
		Metadata:   metadataPayload(),
	}

	_ = events.AppendEvent(db, event)
}

func commandLabel(cmd *cobra.Command) string {
	if cmd == nil {
		return executedCommand
	}

	path := strings.TrimSpace(cmd.CommandPath())
	path = strings.TrimPrefix(path, "carabiner ")
	if path == "carabiner" || path == "" {
		if executedCommand != "" {
			return executedCommand
		}
		return cmd.Name()
	}
	return path
}

func argsPayload(cmd *cobra.Command) string {
	flags := map[string]string{}
	if cmd != nil {
		cmd.Flags().Visit(func(f *pflag.Flag) {
			flags[f.Name] = f.Value.String()
		})
		cmd.InheritedFlags().Visit(func(f *pflag.Flag) {
			flags[f.Name] = f.Value.String()
		})
	}

	payload := map[string]any{
		"argv":       os.Args[1:],
		"positional": []string{},
		"flags":      flags,
	}
	if cmd != nil {
		payload["positional"] = cmd.Flags().Args()
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func metadataPayload() string {
	meta := map[string]any{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"pid":  os.Getpid(),
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func gitValue(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
