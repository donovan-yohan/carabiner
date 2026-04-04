package dossier

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/donovan-yohan/carabiner/internal/carabiner"
	"github.com/donovan-yohan/carabiner/internal/carabiner/agentlytics"
	"github.com/donovan-yohan/carabiner/internal/carabiner/git"
	"github.com/donovan-yohan/carabiner/internal/carabiner/gitai"
)

// Builder assembles a forensic Dossier by joining git blame, git-ai notes,
// and agentlytics session data.
type Builder struct {
	AgentlyticsPath string
	MaxDepth        int
}

// NewBuilder creates a Builder with sensible defaults.
func NewBuilder(agentlyticsPath string) *Builder {
	return &Builder{
		AgentlyticsPath: agentlyticsPath,
		MaxDepth:        10,
	}
}

// Build produces a Dossier for the given file and line.
// The join algorithm:
//  1. git blame → commit SHA
//  2. git notes show --ref=ai <commit> → parse git-ai note
//  3. FindSession in note for file:line → agent_id.id (conversation UUID)
//  4. Query agentlytics chats table by conversation UUID → session metadata
//  5. Assemble dossier with confidence per hop
//  6. (optional) git log --follow for subsequent touches
func (b *Builder) Build(file string, line int, rev string) (*carabiner.Dossier, error) {
	d := &carabiner.Dossier{
		File: file,
		Line: line,
		Rev:  rev,
	}

	// Step 1: git blame
	blame, err := git.Blame(file, line, rev)
	if err != nil {
		return nil, fmt.Errorf("blame: %w", err)
	}
	d.Blame = blame
	d.Hops = append(d.Hops, carabiner.Hop{
		Name:       "line_to_commit",
		Confidence: carabiner.ConfidenceHigh,
		Detail:     fmt.Sprintf("git blame (commit %s)", shortSHA(blame.CommitSHA)),
	})

	// Step 2: read git-ai note
	rawNote, err := git.ShowNote("ai", blame.CommitSHA)
	if err != nil {
		// No note = human-authored or pre-git-ai commit
		d.Hops = append(d.Hops, carabiner.Hop{
			Name:       "commit_to_session",
			Confidence: carabiner.ConfidenceMissing,
			Detail:     err.Error(),
		})
		d.OverallConfidence = carabiner.ConfidenceMissing
		return d, nil
	}

	// Step 3: parse note and find session for this file:line
	note, err := gitai.ParseNote(rawNote)
	if err != nil {
		d.Hops = append(d.Hops, carabiner.Hop{
			Name:       "commit_to_session",
			Confidence: carabiner.ConfidenceMissing,
			Detail:     fmt.Sprintf("malformed git-ai note: %v", err),
		})
		d.OverallConfidence = carabiner.ConfidenceMissing
		return d, nil
	}

	// Normalize file path for matching against attestation
	normalizedFile := normalizePath(file)
	hash, prompt, err := gitai.FindSession(note, normalizedFile, line)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}

	if hash == "" {
		// Line exists in a commit with git-ai notes but isn't in any
		// attested range. Could be a human-edited line within an
		// AI-authored commit.
		d.Hops = append(d.Hops, carabiner.Hop{
			Name:       "commit_to_session",
			Confidence: carabiner.ConfidenceMissing,
			Detail:     "line not in any git-ai attested range",
		})
		d.OverallConfidence = carabiner.ConfidenceMissing
		return d, nil
	}

	// We have a session hash match
	if prompt == nil {
		d.Hops = append(d.Hops, carabiner.Hop{
			Name:       "commit_to_session",
			Confidence: carabiner.ConfidenceMissing,
			Detail:     fmt.Sprintf("session hash %s found in attestation but missing from metadata", hash),
		})
		d.OverallConfidence = carabiner.ConfidenceMissing
		return d, nil
	}

	d.Hops = append(d.Hops, carabiner.Hop{
		Name:       "commit_to_session",
		Confidence: carabiner.ConfidenceHigh,
		Detail:     fmt.Sprintf("git-ai note (session %s)", hash),
	})

	// Build session info from git-ai metadata
	session := &carabiner.SessionInfo{
		ID:    prompt.AgentID.ID,
		Tool:  prompt.AgentID.Tool,
		Model: prompt.AgentID.Model,
	}

	// Step 4: enrich from agentlytics
	if b.AgentlyticsPath != "" {
		aSession, err := agentlytics.QuerySession(b.AgentlyticsPath, prompt.AgentID.ID)
		if err != nil {
			d.Hops = append(d.Hops, carabiner.Hop{
				Name:       "session_to_transcript",
				Confidence: carabiner.ConfidenceMissing,
				Detail:     fmt.Sprintf("agentlytics query failed: %v", err),
			})
		} else if aSession == nil {
			d.Hops = append(d.Hops, carabiner.Hop{
				Name:       "session_to_transcript",
				Confidence: carabiner.ConfidenceMissing,
				Detail:     "session not found in agentlytics cache",
			})
		} else {
			session.Name = aSession.Name
			session.Source = aSession.Source
			session.StartedAt = aSession.CreatedAt
			session.EndedAt = aSession.LastUpdatedAt
			d.Hops = append(d.Hops, carabiner.Hop{
				Name:       "session_to_transcript",
				Confidence: carabiner.ConfidenceHigh,
				Detail:     fmt.Sprintf("agentlytics match (session %q)", aSession.Name),
			})
		}
	}

	d.Session = session

	// Compute overall confidence: weakest link
	d.OverallConfidence = carabiner.ConfidenceHigh
	for _, hop := range d.Hops {
		if hop.Confidence == carabiner.ConfidenceMissing {
			d.OverallConfidence = carabiner.ConfidenceMissing
			break
		}
	}

	return d, nil
}

// shortSHA returns the first 7 characters of a SHA.
func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// normalizePath cleans up a file path for comparison against git-ai attestations.
func normalizePath(file string) string {
	file = filepath.Clean(file)
	file = strings.TrimPrefix(file, "./")
	return file
}
