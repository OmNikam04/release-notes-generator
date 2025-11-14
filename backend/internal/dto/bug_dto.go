package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
)

// ReleaseNoteResponse represents a simple release note in bug responses
type ReleaseNoteResponse struct {
	ID      uuid.UUID `json:"id"`
	Content string    `json:"content"`
	Status  string    `json:"status"`
	Version int       `json:"version"`
}

// ToReleaseNoteResponse converts ReleaseNote model to simple response
func ToReleaseNoteResponse(note *models.ReleaseNote) *ReleaseNoteResponse {
	if note == nil {
		return nil
	}
	return &ReleaseNoteResponse{
		ID:      note.ID,
		Content: note.Content,
		Status:  note.Status,
		Version: note.Version,
	}
}

// BugResponse represents a bug in API responses
type BugResponse struct {
	ID            uuid.UUID            `json:"id"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	BugsbyID      string               `json:"bugsby_id"`
	BugsbyURL     string               `json:"bugsby_url"`
	Title         string               `json:"title"`
	Description   *string              `json:"description"`
	Severity      string               `json:"severity"`
	Priority      string               `json:"priority"`
	BugType       string               `json:"bug_type"`
	CVENumber     *string              `json:"cve_number"`
	AssignedTo    *uuid.UUID           `json:"assigned_to"`
	AssigneeEmail *string              `json:"assignee_email,omitempty"` // Email of assigned user
	ManagerID     *uuid.UUID           `json:"manager_id"`
	ManagerEmail  *string              `json:"manager_email,omitempty"` // Email of manager
	Release       string               `json:"release"`
	Component     string               `json:"component"`
	Status        string               `json:"status"`
	LastSyncedAt  *time.Time           `json:"last_synced_at"`
	SyncStatus    string               `json:"sync_status"`
	ReleaseNote   *ReleaseNoteResponse `json:"release_note,omitempty"`
}

// BugListResponse represents a paginated list of bugs
type BugListResponse struct {
	Bugs       []BugResponse `json:"bugs"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// SyncReleaseRequest represents a request to sync bugs for a release
type SyncReleaseRequest struct {
	Release   string   `json:"release" validate:"required"`
	Status    string   `json:"status,omitempty"`
	Severity  []string `json:"severity,omitempty"`
	BugType   string   `json:"bug_type,omitempty"`
	Component string   `json:"component,omitempty"`
}

// SyncBugByIDRequest represents a request to sync a single bug
type SyncBugByIDRequest struct {
	BugsbyID int `json:"bugsby_id" validate:"required,min=1"`
}

// SyncByQueryRequest represents a request to sync bugs using a custom Bugsby query
type SyncByQueryRequest struct {
	Query string `json:"query" validate:"required"`
	Limit int    `json:"limit,omitempty"` // Optional, defaults to 100
}

// SyncResultResponse represents the result of a sync operation
type SyncResultResponse struct {
	TotalFetched int           `json:"total_fetched"`
	NewBugs      int           `json:"new_bugs"`
	UpdatedBugs  int           `json:"updated_bugs"`
	FailedBugs   int           `json:"failed_bugs"`
	SyncedAt     time.Time     `json:"synced_at"`
	Errors       []string      `json:"errors,omitempty"`
	SyncedBugs   []BugResponse `json:"synced_bugs,omitempty"` // Full bug details for UI display
}

// SyncStatusResponse represents the sync status for a release
type SyncStatusResponse struct {
	Release      string     `json:"release"`
	TotalBugs    int        `json:"total_bugs"`
	SyncedBugs   int        `json:"synced_bugs"`
	PendingBugs  int        `json:"pending_bugs"`
	FailedBugs   int        `json:"failed_bugs"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
}

// UpdateBugRequest represents a request to update a bug
type UpdateBugRequest struct {
	Status     *string    `json:"status,omitempty"`
	AssignedTo *uuid.UUID `json:"assigned_to,omitempty"`
	ManagerID  *uuid.UUID `json:"manager_id,omitempty"`
}

// BugFiltersRequest represents filter parameters for listing bugs
type BugFiltersRequest struct {
	Release        string   `query:"release"`
	Status         []string `query:"status"`
	AssignedTo     string   `query:"assigned_to"` // UUID as string
	ManagerID      string   `query:"manager_id"`  // UUID as string
	Severity       []string `query:"severity"`
	BugType        []string `query:"bug_type"`
	Component      string   `query:"component"`
	HasReleaseNote *bool    `query:"has_release_note"`
	Page           int      `query:"page"`
	Limit          int      `query:"limit"`
	SortBy         string   `query:"sort_by"`
	SortOrder      string   `query:"sort_order"`
}

// ToBugResponse converts a Bug model to BugResponse DTO
func ToBugResponse(bug *models.Bug) *BugResponse {
	if bug == nil {
		return nil
	}

	response := &BugResponse{
		ID:           bug.ID,
		CreatedAt:    bug.CreatedAt,
		UpdatedAt:    bug.UpdatedAt,
		BugsbyID:     bug.BugsbyID,
		BugsbyURL:    bug.BugsbyURL,
		Title:        bug.Title,
		Description:  bug.Description,
		Severity:     bug.Severity,
		Priority:     bug.Priority,
		BugType:      bug.BugType,
		CVENumber:    bug.CVENumber,
		AssignedTo:   bug.AssignedTo,
		ManagerID:    bug.ManagerID,
		Release:      bug.Release,
		Component:    bug.Component,
		Status:       bug.Status,
		LastSyncedAt: bug.LastSyncedAt,
		SyncStatus:   bug.SyncStatus,
	}

	// Include release note if present
	if bug.ReleaseNote != nil {
		response.ReleaseNote = ToReleaseNoteResponse(bug.ReleaseNote)
	}

	return response
}

// ToBugListResponse converts a slice of Bug models to BugListResponse DTO
func ToBugListResponse(bugs []*models.Bug, total int64, page, limit int) *BugListResponse {
	bugResponses := make([]BugResponse, 0, len(bugs))
	for _, bug := range bugs {
		if response := ToBugResponse(bug); response != nil {
			bugResponses = append(bugResponses, *response)
		}
	}

	// Ensure limit is at least 1 to avoid division by zero
	if limit < 1 {
		limit = 20 // Default limit
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &BugListResponse{
		Bugs:       bugResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
