package carabiner

import (
	"fmt"
	"sort"
	"strings"
)

// Check loads learnings and returns those matching the given file paths.
func Check(configDir string, files []string, maxResults int) ([]Learning, error) {
	learnings, err := LoadLearnings(configDir)
	if err != nil {
		return nil, fmt.Errorf("loading learnings: %w", err)
	}

	var matched []Learning
	for _, l := range learnings {
		if matchesAnyPath(l.Paths, files) {
			matched = append(matched, l)
		}
	}

	// Sort by created descending (newest first)
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Created.After(matched[j].Created)
	})

	if maxResults > 0 && len(matched) > maxResults {
		matched = matched[:maxResults]
	}
	return matched, nil
}

// matchesAnyPath returns true if any learning path is a prefix of any input file.
func matchesAnyPath(learningPaths []string, inputFiles []string) bool {
	for _, lp := range learningPaths {
		lp = ensureTrailingSlash(lp)
		for _, f := range inputFiles {
			if strings.HasPrefix(f, lp) {
				return true
			}
		}
	}
	return false
}

// FormatMarkdown renders matched learnings as prompt-ready markdown.
func FormatMarkdown(learnings []Learning) string {
	if len(learnings) == 0 {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Carabiner Quality Patterns (%d relevant)\n\n", len(learnings))

	for i, l := range learnings {
		fmt.Fprintf(&b, "### %s\n", l.Pattern)
		fmt.Fprintf(&b, "**Recommendation:** %s\n", l.Recommendation)
		fmt.Fprintf(&b, "**Source:** %s", l.Source)
		if len(l.Tags) > 0 {
			fmt.Fprintf(&b, " | **Tags:** %s", strings.Join(l.Tags, ", "))
		}
		b.WriteByte('\n')
		if i < len(learnings)-1 {
			b.WriteString("\n---\n\n")
		}
	}
	return b.String()
}
