package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/models"
)

// ===== Request DTOs =====

// GetPendingBugsRequest represents query parameters for getting bugs without release notes
type GetPendingBugsRequest struct {
	AssignedToMe bool     `query:"assigned_to_me"` // Filter by current user
	Release      string   `query:"release"`
	Status       []string `query:"status"`
	Severity     []string `query:"severity"`
	Component    string   `query:"component"`
	Page         int      `query:"page"`
	Limit        int      `query:"limit"`
	SortBy       string   `query:"sort_by"`
	SortOrder    string   `query:"sort_order"`
}

// GenerateReleaseNoteRequest represents a request to generate a release note
type GenerateReleaseNoteRequest struct {
	BugID         uuid.UUID `json:"bug_id" validate:"required"`
	ManualContent *string   `json:"manual_content,omitempty"` // Optional manual content
}

// UpdateReleaseNoteRequest represents a request to update a release note
type UpdateReleaseNoteRequest struct {
	Content string `json:"content" validate:"required"`
	Status  string `json:"status,omitempty" validate:"omitempty,oneof=draft ai_generated dev_approved mgr_approved rejected"`
}

// BulkGenerateRequest represents a request to generate multiple release notes
type BulkGenerateRequest struct {
	BugIDs  []uuid.UUID `json:"bug_ids" validate:"required,min=1"`
	Release string      `json:"release,omitempty"` // Optional: generate for all bugs in a release
}

// ApproveReleaseNoteRequest represents a request to approve/reject a release note
type ApproveReleaseNoteRequest struct {
	Action   string  `json:"action" validate:"required,oneof=approve reject"`
	Feedback *string `json:"feedback,omitempty"`
}

// ===== Response DTOs =====

// CommitInfoResponse represents parsed commit information
type CommitInfoResponse struct {
	CommitHash  string    `json:"commit_hash"`
	GerritURL   string    `json:"gerrit_url"`
	Repository  string    `json:"repository"`
	Branch      string    `json:"branch"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	ChangeID    string    `json:"change_id"`
	MergedBy    string    `json:"merged_by"`
	CommentID   int       `json:"comment_id"`
	CommentedAt time.Time `json:"commented_at"`
}

// BugContextResponse represents bug details with commit information for AI generation
type BugContextResponse struct {
	Bug              *BugResponse          `json:"bug"`
	Comments         []CommitInfoResponse  `json:"comments"`
	CommitCount      int                   `json:"commit_count"`
	ReadyForGenerate bool                  `json:"ready_for_generation"`
}

// ReleaseNoteDetailResponse represents a detailed release note response
type ReleaseNoteDetailResponse struct {
	ID                uuid.UUID  `json:"id"`
	BugID             uuid.UUID  `json:"bug_id"`
	Content           string     `json:"content"`
	Version           int        `json:"version"`
	GeneratedBy       string     `json:"generated_by"`
	AIModel           *string    `json:"ai_model,omitempty"`
	AIConfidence      *float64   `json:"ai_confidence,omitempty"`
	Status            string     `json:"status"`
	CreatedByID       *uuid.UUID `json:"created_by_id,omitempty"`
	ApprovedByDevID   *uuid.UUID `json:"approved_by_dev_id,omitempty"`
	ApprovedByMgrID   *uuid.UUID `json:"approved_by_mgr_id,omitempty"`
	DevApprovedAt     *time.Time `json:"dev_approved_at,omitempty"`
	MgrApprovedAt     *time.Time `json:"mgr_approved_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	Bug               *BugResponse `json:"bug,omitempty"`
}

// PendingBugsResponse represents a list of bugs without release notes
type PendingBugsResponse struct {
	Bugs       []BugResponse `json:"bugs"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// BulkGenerateItemResponse represents the result of generating one release note
type BulkGenerateItemResponse struct {
	BugID         uuid.UUID  `json:"bug_id"`
	ReleaseNoteID *uuid.UUID `json:"release_note_id,omitempty"`
	Status        string     `json:"status"` // "success" or "failed"
	Error         *string    `json:"error,omitempty"`
}

// BulkGenerateResponse represents the result of bulk generation
type BulkGenerateResponse struct {
	Total     int                            `json:"total"`
	Generated int                            `json:"generated"`
	Failed    int                            `json:"failed"`
	Results   []BulkGenerateItemResponse     `json:"results"`
}

// ===== Converter Functions =====

// ToCommitInfoResponse converts ParsedCommitInfo to CommitInfoResponse
func ToCommitInfoResponse(info *bugsby.ParsedCommitInfo) *CommitInfoResponse {
	if info == nil {
		return nil
	}
	return &CommitInfoResponse{
		CommitHash:  info.CommitHash,
		GerritURL:   info.GerritURL,
		Repository:  info.Repository,
		Branch:      info.Branch,
		Title:       info.Title,
		Message:     info.Message,
		ChangeID:    info.ChangeID,
		MergedBy:    info.MergedBy,
		CommentID:   info.CommentID,
		CommentedAt: info.CommentedAt,
	}
}

// ToReleaseNoteDetailResponse converts ReleaseNote model to detailed response
func ToReleaseNoteDetailResponse(note *models.ReleaseNote) *ReleaseNoteDetailResponse {
	if note == nil {
		return nil
	}

	response := &ReleaseNoteDetailResponse{
		ID:              note.ID,
		BugID:           note.BugID,
		Content:         note.Content,
		Version:         note.Version,
		GeneratedBy:     note.GeneratedBy,
		AIModel:         note.AIModel,
		AIConfidence:    note.AIConfidence,
		Status:          note.Status,
		CreatedByID:     note.CreatedByID,
		ApprovedByDevID: note.ApprovedByDevID,
		ApprovedByMgrID: note.ApprovedByMgrID,
		DevApprovedAt:   note.DevApprovedAt,
		MgrApprovedAt:   note.MgrApprovedAt,
		CreatedAt:       note.CreatedAt,
		UpdatedAt:       note.UpdatedAt,
	}

	// Include bug if preloaded
	if note.Bug != nil {
		response.Bug = ToBugResponse(note.Bug)
	}

	return response
}

