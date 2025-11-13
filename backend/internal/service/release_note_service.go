package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"github.com/omnikam04/release-notes-generator/internal/repository"
	"gorm.io/gorm"
)

// ReleaseNoteService defines the interface for release note business logic
type ReleaseNoteService interface {
	// Get bugs without release notes
	GetPendingBugs(ctx context.Context, userID uuid.UUID, filters *PendingBugsFilters, pagination *repository.Pagination) (*PendingBugsResult, error)

	// Get bugs WITH release notes (Kanban view)
	GetReleaseNotes(ctx context.Context, userID uuid.UUID, filters *ReleaseNotesFilters, pagination *repository.Pagination) (*ReleaseNotesResult, error)

	// Get bug context for AI generation
	GetBugContext(ctx context.Context, bugID uuid.UUID) (*BugContext, error)

	// Generate release note (placeholder for now, AI later)
	GenerateReleaseNote(ctx context.Context, bugID uuid.UUID, userID uuid.UUID, manualContent *string) (*models.ReleaseNote, error)

	// Bulk generate release notes
	BulkGenerateReleaseNotes(ctx context.Context, bugIDs []uuid.UUID, userID uuid.UUID) (*BulkGenerateResult, error)

	// Update release note
	UpdateReleaseNote(ctx context.Context, id uuid.UUID, content string, status string, userID uuid.UUID) (*models.ReleaseNote, error)

	// Get release note by bug ID
	GetReleaseNoteByBugID(ctx context.Context, bugID uuid.UUID) (*models.ReleaseNote, error)

	// Approve/Reject release note (manager)
	ApproveReleaseNote(ctx context.Context, id uuid.UUID, managerID uuid.UUID, feedback *string) error
	RejectReleaseNote(ctx context.Context, id uuid.UUID, managerID uuid.UUID, feedback string) error
}

// PendingBugsFilters represents filters for pending bugs query
type PendingBugsFilters struct {
	AssignedTo *uuid.UUID
	ManagerID  *uuid.UUID
	Release    string
	Status     []string
	Severity   []string
	Component  string
}

// ReleaseNotesFilters represents filters for release notes query (bugs WITH release notes)
type ReleaseNotesFilters struct {
	AssignedTo *uuid.UUID // Filter by bug's assigned developer
	ManagerID  *uuid.UUID // Filter by bug's manager
	Status     []string   // Filter by release note status
	Release    string     // Filter by bug's release
	Component  string     // Filter by bug's component
}

// BugContext represents bug details with commit information
type BugContext struct {
	Bug         *models.Bug
	Comments    []*bugsby.ParsedCommitInfo
	CommitCount int
}

// PendingBugsResult represents the result of pending bugs query
type PendingBugsResult struct {
	Bugs       []*models.Bug
	Total      int64
	Pagination *repository.Pagination
}

// ReleaseNotesResult represents the result of release notes query (bugs WITH release notes)
type ReleaseNotesResult struct {
	ReleaseNotes []*models.ReleaseNote
	Total        int64
	Pagination   *repository.Pagination
}

// BulkGenerateResult represents the result of bulk generation
type BulkGenerateResult struct {
	Total     int
	Generated int
	Failed    int
	Results   []BulkGenerateItem
}

// BulkGenerateItem represents the result of generating one release note
type BulkGenerateItem struct {
	BugID         uuid.UUID
	ReleaseNoteID *uuid.UUID
	Status        string
	Error         *string
}

// releaseNoteService is the concrete implementation
type releaseNoteService struct {
	releaseNoteRepo repository.ReleaseNoteRepository
	bugRepo         repository.BugRepository
	bugsbyClient    bugsby.Client
	aiService       AIService
	db              *gorm.DB
}

// NewReleaseNoteService creates a new release note service instance
func NewReleaseNoteService(
	releaseNoteRepo repository.ReleaseNoteRepository,
	bugRepo repository.BugRepository,
	bugsbyClient bugsby.Client,
	aiService AIService,
	db *gorm.DB,
) ReleaseNoteService {
	return &releaseNoteService{
		releaseNoteRepo: releaseNoteRepo,
		bugRepo:         bugRepo,
		bugsbyClient:    bugsbyClient,
		aiService:       aiService,
		db:              db,
	}
}

// GetPendingBugs retrieves bugs that don't have release notes yet
func (s *releaseNoteService) GetPendingBugs(
	ctx context.Context,
	userID uuid.UUID,
	filters *PendingBugsFilters,
	pagination *repository.Pagination,
) (*PendingBugsResult, error) {
	// Convert to repository filters
	repoFilters := &repository.PendingBugsFilters{
		AssignedTo: filters.AssignedTo,
		ManagerID:  filters.ManagerID,
		Release:    filters.Release,
		Status:     filters.Status,
		Severity:   filters.Severity,
		Component:  filters.Component,
	}

	// If no specific user filter, default to current user
	if repoFilters.AssignedTo == nil && repoFilters.ManagerID == nil {
		repoFilters.AssignedTo = &userID
	}

	bugs, total, err := s.releaseNoteRepo.ListPendingBugs(repoFilters, pagination)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get pending bugs")
		return nil, fmt.Errorf("failed to get pending bugs: %w", err)
	}

	logger.Info().
		Str("user_id", userID.String()).
		Int64("total", total).
		Int("returned", len(bugs)).
		Msg("Retrieved pending bugs")

	return &PendingBugsResult{
		Bugs:       bugs,
		Total:      total,
		Pagination: pagination,
	}, nil
}

// GetReleaseNotes retrieves bugs WITH release notes (Kanban view)
func (s *releaseNoteService) GetReleaseNotes(
	ctx context.Context,
	userID uuid.UUID,
	filters *ReleaseNotesFilters,
	pagination *repository.Pagination,
) (*ReleaseNotesResult, error) {
	// Convert to repository filters
	repoFilters := &repository.ReleaseNoteFilters{
		AssignedTo: filters.AssignedTo,
		ManagerID:  filters.ManagerID,
		Status:     filters.Status,
		Release:    filters.Release,
		Component:  filters.Component,
	}

	// Get release notes
	notes, total, err := s.releaseNoteRepo.List(repoFilters, pagination)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get release notes")
		return nil, fmt.Errorf("failed to get release notes: %w", err)
	}

	logger.Info().
		Str("user_id", userID.String()).
		Int64("total", total).
		Int("returned", len(notes)).
		Msg("Retrieved release notes")

	return &ReleaseNotesResult{
		ReleaseNotes: notes,
		Total:        total,
		Pagination:   pagination,
	}, nil
}

// GetBugContext retrieves bug details with commit information from Bugsby
func (s *releaseNoteService) GetBugContext(ctx context.Context, bugID uuid.UUID) (*BugContext, error) {
	// Get bug from database
	bug, err := s.bugRepo.FindByID(bugID)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", bugID.String()).Msg("Bug not found")
		return nil, fmt.Errorf("bug not found: %w", err)
	}

	// Parse Bugsby ID
	bugsbyID := 0
	if _, err := fmt.Sscanf(bug.BugsbyID, "%d", &bugsbyID); err != nil {
		logger.Error().Err(err).Str("bugsby_id", bug.BugsbyID).Msg("Invalid Bugsby ID")
		return nil, fmt.Errorf("invalid bugsby ID: %w", err)
	}

	// Fetch comments from Bugsby (filtered by gerrit@arista.com)
	commentsResp, err := s.bugsbyClient.GetBugCommentsFiltered(ctx, bugsbyID, "gerrit@arista.com")
	if err != nil {
		logger.Error().Err(err).Int("bugsby_id", bugsbyID).Msg("Failed to fetch comments from Bugsby")
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Parse commit information from comments
	var parsedCommits []*bugsby.ParsedCommitInfo
	for i := range commentsResp.Comments {
		if parsed := s.bugsbyClient.ParseCommitInfo(&commentsResp.Comments[i]); parsed != nil {
			parsedCommits = append(parsedCommits, parsed)
		}
	}

	logger.Info().
		Str("bug_id", bugID.String()).
		Int("bugsby_id", bugsbyID).
		Int("total_comments", len(commentsResp.Comments)).
		Int("parsed_commits", len(parsedCommits)).
		Msg("Retrieved bug context")

	return &BugContext{
		Bug:         bug,
		Comments:    parsedCommits,
		CommitCount: len(parsedCommits),
	}, nil
}

// GenerateReleaseNote generates a release note for a bug
// Phase 1: Creates a placeholder/template
// Phase 2: Will integrate with AI service
func (s *releaseNoteService) GenerateReleaseNote(
	ctx context.Context,
	bugID uuid.UUID,
	userID uuid.UUID,
	manualContent *string,
) (*models.ReleaseNote, error) {
	// Check if release note already exists
	existing, err := s.releaseNoteRepo.FindByBugID(bugID)
	if err == nil && existing != nil {
		logger.Warn().Str("bug_id", bugID.String()).Msg("Release note already exists")
		return nil, fmt.Errorf("release note already exists for this bug")
	}

	// Get bug details
	bug, err := s.bugRepo.FindByID(bugID)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", bugID.String()).Msg("Bug not found")
		return nil, fmt.Errorf("bug not found: %w", err)
	}

	var content string
	var generatedBy string
	var aiModel *string
	var aiConfidence *float64
	var status string

	if manualContent != nil && *manualContent != "" {
		// Use manual content
		content = *manualContent
		generatedBy = "manual"
		status = "draft"
	} else {
		// Try AI generation first
		if s.aiService != nil {
			// Get bug context (commits)
			bugContext, err := s.GetBugContext(ctx, bugID)
			if err != nil {
				logger.Warn().Err(err).Str("bug_id", bugID.String()).Msg("Failed to get bug context, will try AI without commits")
			}

			// Generate with AI
			aiContent, confidence, aiErr := s.aiService.GenerateReleaseNote(ctx, bug, bugContext.Comments)
			if aiErr == nil && aiContent != "" {
				// AI generation successful
				content = aiContent
				generatedBy = "ai"
				status = "ai_generated"
				modelName := "gemini-2.5-pro" // Get from config
				aiModel = &modelName
				aiConfidence = &confidence

				logger.Info().
					Str("bug_id", bugID.String()).
					Float64("confidence", confidence).
					Msg("Successfully generated release note with AI")
			} else {
				// AI generation failed, fallback to placeholder
				logger.Warn().
					Err(aiErr).
					Str("bug_id", bugID.String()).
					Msg("AI generation failed, falling back to placeholder")
				content = s.generatePlaceholderContent(bug)
				generatedBy = "placeholder"
				status = "draft"
			}
		} else {
			// No AI service available, use placeholder
			logger.Warn().Str("bug_id", bugID.String()).Msg("AI service not available, using placeholder")
			content = s.generatePlaceholderContent(bug)
			generatedBy = "placeholder"
			status = "draft"
		}
	}

	// Create release note
	note := &models.ReleaseNote{
		ID:           uuid.New(),
		BugID:        bugID,
		Content:      content,
		Version:      1,
		GeneratedBy:  generatedBy,
		AIModel:      aiModel,
		AIConfidence: aiConfidence,
		Status:       status,
		CreatedByID:  &userID,
	}

	// Save to database
	if err := s.releaseNoteRepo.Create(note); err != nil {
		logger.Error().Err(err).Str("bug_id", bugID.String()).Msg("Failed to create release note")
		return nil, fmt.Errorf("failed to create release note: %w", err)
	}

	// Update bug status
	bug.Status = "ai_generated"
	if err := s.bugRepo.Update(bug); err != nil {
		logger.Error().Err(err).Str("bug_id", bugID.String()).Msg("Failed to update bug status")
		// Don't fail the operation, just log the error
	}

	logger.Info().
		Str("bug_id", bugID.String()).
		Str("note_id", note.ID.String()).
		Str("generated_by", generatedBy).
		Msg("Release note created")

	return note, nil
}

// generatePlaceholderContent creates a template release note
func (s *releaseNoteService) generatePlaceholderContent(bug *models.Bug) string {
	var builder strings.Builder

	// Format: "Fixed [title] in [component]"
	builder.WriteString("Fixed: ")
	builder.WriteString(bug.Title)

	if bug.Component != "" {
		builder.WriteString(" in ")
		builder.WriteString(bug.Component)
	}

	builder.WriteString(".\n\n")
	builder.WriteString("Severity: ")
	builder.WriteString(bug.Severity)
	builder.WriteString("\n")
	builder.WriteString("Priority: ")
	builder.WriteString(bug.Priority)
	builder.WriteString("\n\n")
	builder.WriteString("[This is a placeholder release note. Please edit with actual details.]")

	return builder.String()
}

// UpdateReleaseNote updates an existing release note
func (s *releaseNoteService) UpdateReleaseNote(
	ctx context.Context,
	id uuid.UUID,
	content string,
	status string,
	userID uuid.UUID,
) (*models.ReleaseNote, error) {
	// Get existing note
	note, err := s.releaseNoteRepo.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("note_id", id.String()).Msg("Release note not found")
		return nil, fmt.Errorf("release note not found: %w", err)
	}

	// Update fields
	note.Content = content
	note.Version++

	if status != "" {
		note.Status = status

		// Set approval fields based on status
		if status == "dev_approved" {
			now := time.Now()
			note.ApprovedByDevID = &userID
			note.DevApprovedAt = &now

			// Update bug status
			if note.Bug != nil {
				note.Bug.Status = "dev_approved"
				if err := s.bugRepo.Update(note.Bug); err != nil {
					logger.Error().Err(err).Msg("Failed to update bug status")
				}
			}
		}
	}

	// Save changes
	if err := s.releaseNoteRepo.Update(note); err != nil {
		logger.Error().Err(err).Str("note_id", id.String()).Msg("Failed to update release note")
		return nil, fmt.Errorf("failed to update release note: %w", err)
	}

	logger.Info().
		Str("note_id", id.String()).
		Str("status", status).
		Int("version", note.Version).
		Msg("Release note updated")

	return note, nil
}

// GetReleaseNoteByBugID retrieves a release note by bug ID
func (s *releaseNoteService) GetReleaseNoteByBugID(ctx context.Context, bugID uuid.UUID) (*models.ReleaseNote, error) {
	note, err := s.releaseNoteRepo.FindByBugID(bugID)
	if err != nil {
		logger.Error().Err(err).Str("bug_id", bugID.String()).Msg("Release note not found")
		return nil, fmt.Errorf("release note not found: %w", err)
	}
	return note, nil
}

// BulkGenerateReleaseNotes generates release notes for multiple bugs
func (s *releaseNoteService) BulkGenerateReleaseNotes(
	ctx context.Context,
	bugIDs []uuid.UUID,
	userID uuid.UUID,
) (*BulkGenerateResult, error) {
	result := &BulkGenerateResult{
		Total:   len(bugIDs),
		Results: make([]BulkGenerateItem, 0, len(bugIDs)),
	}

	for _, bugID := range bugIDs {
		item := BulkGenerateItem{
			BugID:  bugID,
			Status: "success",
		}

		// Try to generate release note
		note, err := s.GenerateReleaseNote(ctx, bugID, userID, nil)
		if err != nil {
			result.Failed++
			item.Status = "failed"
			errMsg := err.Error()
			item.Error = &errMsg
		} else {
			result.Generated++
			item.ReleaseNoteID = &note.ID
		}

		result.Results = append(result.Results, item)
	}

	logger.Info().
		Int("total", result.Total).
		Int("generated", result.Generated).
		Int("failed", result.Failed).
		Msg("Bulk generation completed")

	return result, nil
}

// ApproveReleaseNote approves a release note (manager only)
func (s *releaseNoteService) ApproveReleaseNote(
	ctx context.Context,
	id uuid.UUID,
	managerID uuid.UUID,
	feedback *string,
) error {
	// Get release note
	note, err := s.releaseNoteRepo.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("note_id", id.String()).Msg("Release note not found")
		return fmt.Errorf("release note not found: %w", err)
	}

	// Update status
	now := time.Now()
	note.Status = "mgr_approved"
	note.ApprovedByMgrID = &managerID
	note.MgrApprovedAt = &now

	// Save changes
	if err := s.releaseNoteRepo.Update(note); err != nil {
		logger.Error().Err(err).Str("note_id", id.String()).Msg("Failed to approve release note")
		return fmt.Errorf("failed to approve release note: %w", err)
	}

	// Update bug status
	if note.Bug != nil {
		note.Bug.Status = "mgr_approved"
		if err := s.bugRepo.Update(note.Bug); err != nil {
			logger.Error().Err(err).Msg("Failed to update bug status")
		}
	}

	logger.Info().
		Str("note_id", id.String()).
		Str("manager_id", managerID.String()).
		Msg("Release note approved")

	return nil
}

// RejectReleaseNote rejects a release note (manager only)
func (s *releaseNoteService) RejectReleaseNote(
	ctx context.Context,
	id uuid.UUID,
	managerID uuid.UUID,
	feedback string,
) error {
	// Get release note
	note, err := s.releaseNoteRepo.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("note_id", id.String()).Msg("Release note not found")
		return fmt.Errorf("release note not found: %w", err)
	}

	// Update status
	note.Status = "rejected"

	// Save changes
	if err := s.releaseNoteRepo.Update(note); err != nil {
		logger.Error().Err(err).Str("note_id", id.String()).Msg("Failed to reject release note")
		return fmt.Errorf("failed to reject release note: %w", err)
	}

	// Update bug status
	if note.Bug != nil {
		note.Bug.Status = "rejected"
		if err := s.bugRepo.Update(note.Bug); err != nil {
			logger.Error().Err(err).Msg("Failed to update bug status")
		}
	}

	logger.Info().
		Str("note_id", id.String()).
		Str("manager_id", managerID.String()).
		Str("feedback", feedback).
		Msg("Release note rejected")

	return nil
}
