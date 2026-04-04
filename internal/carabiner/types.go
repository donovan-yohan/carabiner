package carabiner

import "time"

// ConfidenceLevel indicates how trustworthy a single hop in the attribution chain is.
type ConfidenceLevel string

const (
	ConfidenceHigh    ConfidenceLevel = "high"
	ConfidenceMissing ConfidenceLevel = "missing"
)

// Hop represents one step in the attribution chain with its confidence.
type Hop struct {
	Name       string          `json:"name"`
	Confidence ConfidenceLevel `json:"confidence"`
	Detail     string          `json:"detail,omitempty"`
}

// SessionInfo holds metadata about an AI agent session.
type SessionInfo struct {
	ID        string    `json:"id"`
	Tool      string    `json:"tool"`
	Model     string    `json:"model,omitempty"`
	Name      string    `json:"name,omitempty"`
	Source    string    `json:"source,omitempty"`
	StartedAt time.Time `json:"started_at,omitzero"`
	EndedAt   time.Time `json:"ended_at,omitzero"`
}

// BlameResult holds the output of git blame for a single line.
type BlameResult struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	CommitSHA string `json:"commit_sha"`
	Author    string `json:"author"`
	Date      string `json:"date"`
	Content   string `json:"content"`
}

// SubsequentTouch represents a later commit that modified the same line.
type SubsequentTouch struct {
	CommitSHA string       `json:"commit_sha"`
	Author    string       `json:"author"`
	Date      string       `json:"date"`
	Session   *SessionInfo `json:"session,omitempty"`
}

// Dossier is the complete forensic report for a line of code.
type Dossier struct {
	File              string            `json:"file"`
	Line              int               `json:"line"`
	Rev               string            `json:"rev,omitempty"`
	Blame             *BlameResult      `json:"blame"`
	Session           *SessionInfo      `json:"session,omitempty"`
	Hops              []Hop             `json:"hops"`
	OverallConfidence ConfidenceLevel   `json:"overall_confidence"`
	SubsequentTouches []SubsequentTouch `json:"subsequent_touches,omitempty"`
}
