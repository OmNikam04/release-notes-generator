package gemini

import (
	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/models"
)

// GenerateReleaseNoteRequest represents a request to generate a release note
type GenerateReleaseNoteRequest struct {
	Bug     *models.Bug
	Commits []*bugsby.ParsedCommitInfo
}

// GenerateReleaseNoteResponse represents the AI-generated release note
type GenerateReleaseNoteResponse struct {
	Content    string
	Confidence float64
	Model      string
}

// Config holds Gemini client configuration
type Config struct {
	ProjectID string
	Location  string
	Model     string
}
