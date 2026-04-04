package gitai

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Note represents a parsed git-ai note (v3.0.0 format).
// Format: attestation section (file -> session_hash -> line ranges)
// separated by "---" from a JSON metadata section.
type Note struct {
	Attestations []Attestation
	Metadata     Metadata
}

// Attestation represents one file's attribution in the note.
type Attestation struct {
	File     string
	Sessions []SessionAttestation
}

// SessionAttestation maps a session hash to line ranges within a file.
type SessionAttestation struct {
	Hash       string
	LineRanges []LineRange
}

// LineRange is a start-end pair (inclusive).
type LineRange struct {
	Start int
	End   int
}

// Metadata is the JSON section of a git-ai note.
type Metadata struct {
	SchemaVersion string            `json:"schema_version"`
	BaseCommitSHA string            `json:"base_commit_sha"`
	Prompts       map[string]Prompt `json:"prompts"`
}

// Prompt holds agent identity info keyed by session hash in the metadata.
type Prompt struct {
	AgentID  AgentID  `json:"agent_id"`
	Messages []string `json:"messages"`
}

// AgentID identifies the agent that produced code.
type AgentID struct {
	Tool  string `json:"tool"`
	ID    string `json:"id"`
	Model string `json:"model"`
}

// SessionHash computes the git-ai session hash: SHA-256("{tool}:{id}")[:16].
func SessionHash(tool, conversationID string) string {
	h := sha256.Sum256([]byte(tool + ":" + conversationID))
	return fmt.Sprintf("%x", h[:])[:16]
}

// ParseNote parses a raw git-ai note string into structured data.
func ParseNote(raw string) (*Note, error) {
	parts := strings.SplitN(raw, "\n---\n", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("git-ai note missing --- separator between attestation and metadata")
	}

	attestations, err := parseAttestations(parts[0])
	if err != nil {
		return nil, fmt.Errorf("parsing attestations: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("parsing metadata JSON: %w", err)
	}

	return &Note{
		Attestations: attestations,
		Metadata:     meta,
	}, nil
}

// parseAttestations parses the attestation section.
// Format:
//
//	filepath
//	  session_hash line_ranges
//	  session_hash line_ranges
//	filepath
//	  session_hash line_ranges
func parseAttestations(raw string) ([]Attestation, error) {
	var attestations []Attestation
	var current *Attestation

	for _, line := range strings.Split(raw, "\n") {
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "  ") {
			// File path line
			if current != nil {
				attestations = append(attestations, *current)
			}
			current = &Attestation{File: line}
			continue
		}

		// Session line: "  hash ranges"
		if current == nil {
			return nil, fmt.Errorf("session line without file: %q", line)
		}

		trimmed := strings.TrimPrefix(line, "  ")
		parts := strings.SplitN(trimmed, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed session line: %q", line)
		}

		ranges, err := parseLineRanges(parts[1])
		if err != nil {
			return nil, fmt.Errorf("parsing line ranges in %q: %w", line, err)
		}

		current.Sessions = append(current.Sessions, SessionAttestation{
			Hash:       parts[0],
			LineRanges: ranges,
		})
	}

	if current != nil {
		attestations = append(attestations, *current)
	}

	return attestations, nil
}

// parseLineRanges parses "1-5,7-10" into []LineRange.
func parseLineRanges(raw string) ([]LineRange, error) {
	var ranges []LineRange
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		bounds := strings.SplitN(part, "-", 2)
		start, err := strconv.Atoi(bounds[0])
		if err != nil {
			return nil, fmt.Errorf("invalid range start %q: %w", bounds[0], err)
		}

		end := start
		if len(bounds) == 2 {
			end, err = strconv.Atoi(bounds[1])
			if err != nil {
				return nil, fmt.Errorf("invalid range end %q: %w", bounds[1], err)
			}
		}

		ranges = append(ranges, LineRange{Start: start, End: end})
	}
	return ranges, nil
}

// FindSession looks up which session (if any) is attributed for a given
// file and line number in the note. Returns the session hash and the
// corresponding Prompt metadata, or nil if the line is not AI-attributed.
func FindSession(note *Note, file string, line int) (string, *Prompt, error) {
	for _, att := range note.Attestations {
		if att.File != file {
			continue
		}
		for _, sess := range att.Sessions {
			for _, lr := range sess.LineRanges {
				if line >= lr.Start && line <= lr.End {
					if prompt, ok := note.Metadata.Prompts[sess.Hash]; ok {
						return sess.Hash, &prompt, nil
					}
					return sess.Hash, nil, nil
				}
			}
		}
	}
	return "", nil, nil
}
