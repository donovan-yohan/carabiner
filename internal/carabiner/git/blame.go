package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
)

// porcelainLineRe matches the commit hash line in git blame --porcelain output.
var porcelainLineRe = regexp.MustCompile(`^([0-9a-f]{40})\s`)

// Blame runs git blame for a single line and returns the result.
// If rev is empty, blame is computed against the working tree, including uncommitted changes. Uses -C to detect copies/moves.
func Blame(file string, line int, rev string) (*carabiner.BlameResult, error) {
	lineRange := fmt.Sprintf("%d,%d", line, line)
	args := []string{"blame", "--porcelain", "-C", "-L", lineRange}
	if rev != "" {
		args = append(args, rev)
	}
	args = append(args, "--", file)

	out, err := RunGit(args...)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "no such path") || strings.Contains(errStr, "no such ref") {
			return nil, fmt.Errorf("file %q not tracked by git", file)
		}
		return nil, err
	}

	if out == "" {
		return nil, fmt.Errorf("line %d does not exist in %s", line, file)
	}

	return parsePorcelainBlame(out, file, line)
}

// parsePorcelainBlame extracts blame info from git blame --porcelain output.
func parsePorcelainBlame(raw string, file string, line int) (*carabiner.BlameResult, error) {
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty blame output for %s:%d", file, line)
	}

	result := &carabiner.BlameResult{
		File: file,
		Line: line,
	}

	// First line: <sha> <orig-line> <final-line> [<num-lines>]
	match := porcelainLineRe.FindStringSubmatch(lines[0])
	if match == nil {
		return nil, fmt.Errorf("unexpected blame format: %q", lines[0])
	}
	result.CommitSHA = match[1]

	// Parse header fields and content line
	for _, l := range lines[1:] {
		switch {
		case strings.HasPrefix(l, "author "):
			result.Author = strings.TrimPrefix(l, "author ")
		case strings.HasPrefix(l, "author-time "):
			result.Date = strings.TrimPrefix(l, "author-time ")
		case strings.HasPrefix(l, "\t"):
			result.Content = strings.TrimPrefix(l, "\t")
		}
	}

	return result, nil
}
