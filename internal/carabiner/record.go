package carabiner

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Record creates a new learning from gate input and saves it to disk.
func Record(configDir string, gateID string, rawInput string, gateResult *GateResult, skipExtraction bool, extractCmd string) (*Learning, error) {
	now := time.Now().UTC()
	id := uuid.New().String()

	var (
		pattern        string
		recommendation string
		paths          []string
		tags           []string
		source         string
		branch         string
		commit         string
		files          []string
	)

	if gateResult != nil {
		source = "gate:" + gateResult.GateID
		branch = gateResult.Branch
		commit = gateResult.Commit
		files = gateResult.Files
	} else {
		source = "manual"
	}

	if skipExtraction || extractCmd == "" {
		if gateResult != nil {
			pattern = firstLine(gateResult.Rationale)
			recommendation = gateResult.Rationale
			paths = derivePaths(gateResult.Files)
		} else {
			pattern = firstLine(rawInput)
			recommendation = rawInput
		}
	} else {
		rationale := rawInput
		if gateResult != nil {
			rationale = gateResult.Rationale
			files = gateResult.Files
		}

		result, err := Extract(extractCmd, rationale, files, 30*time.Second)
		if err != nil {
			fmt.Printf("warning: extraction failed, storing raw input: %v\n", err)
			pattern = firstLine(rawInput)
			recommendation = rawInput
			if gateResult != nil {
				paths = derivePaths(gateResult.Files)
			}
		} else {
			pattern = result.Pattern
			recommendation = result.Recommendation
			paths = result.Paths
			tags = result.Tags
		}
	}

	learning := &Learning{
		ID:             id,
		Created:        now,
		Source:         source,
		Paths:          normalizePaths(paths),
		Tags:           tags,
		Pattern:        pattern,
		Recommendation: recommendation,
		RawInput:       rawInput,
		Artifacts: LearningArtifacts{
			Branch: branch,
			Commit: commit,
		},
	}

	if err := SaveLearning(configDir, learning); err != nil {
		return nil, fmt.Errorf("saving learning: %w", err)
	}

	signal := &Signal{
		Timestamp:  now,
		LearningID: id,
		GateID:     gateID,
		Result:     "fail",
		Branch:     branch,
		Commit:     commit,
		Files:      files,
	}

	if err := AppendSignal(configDir, signal); err != nil {
		return nil, fmt.Errorf("appending initial signal: %w", err)
	}

	return learning, nil
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

func derivePaths(files []string) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, f := range files {
		dir := filepath.Dir(f)
		if dir == "." {
			continue
		}
		dir = ensureTrailingSlash(dir)
		if !seen[dir] {
			seen[dir] = true
			paths = append(paths, dir)
		}
	}
	return paths
}

func normalizePaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		if p != "" {
			result = append(result, ensureTrailingSlash(p))
		}
	}
	return result
}

func ensureTrailingSlash(p string) string {
	if !strings.HasSuffix(p, "/") {
		return p + "/"
	}
	return p
}
