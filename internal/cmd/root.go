package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
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
	Short: "Agent-agnostic harness for coding agents",
	Long:  "Carabiner is a repo's institutional memory for coding agents. It persists quality patterns across sessions and surfaces them when agents touch affected files.",
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
		logEvent(cmd, true)
	},
}

func Execute() {
	executed, err := rootCmd.ExecuteC()
	if err != nil {
		exitCode = 2
		if executed != nil {
			executedCommand = commandLabel(executed)
		}
		logEvent(executed, false)
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

func logEvent(cmd *cobra.Command, async bool) {
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

	if async {
		go func() {
			_ = events.AppendEvent(db, event)
		}()
		return
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
	username := ""
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	meta := map[string]any{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"user":       username,
		"pid":        os.Getpid(),
		"cwd":        safeGetwd(),
		"commandRaw": os.Args,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func safeGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

func gitValue(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
