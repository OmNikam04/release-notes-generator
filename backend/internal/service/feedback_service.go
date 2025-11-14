package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"github.com/omnikam04/release-notes-generator/internal/repository"
)

// FeedbackService handles feedback capture and management
type FeedbackService interface {
	// Capture feedback when manager approves with corrections
	CaptureFeedback(ctx context.Context, req *CaptureFeedbackRequest) (*models.Feedback, error)

	// Get feedback
	GetFeedback(ctx context.Context, id uuid.UUID) (*models.Feedback, error)
	GetFeedbackByReleaseNote(ctx context.Context, releaseNoteID uuid.UUID) (*models.Feedback, error)
	GetManagerFeedback(ctx context.Context, managerID uuid.UUID, page, limit int) ([]*models.Feedback, int64, error)

	// Update effectiveness score
	UpdateEffectivenessScore(ctx context.Context, feedbackID uuid.UUID, score float64) error
	IncrementUsageCount(ctx context.Context, feedbackID uuid.UUID) error
}

// CaptureFeedbackRequest represents a request to capture manager feedback
type CaptureFeedbackRequest struct {
	ReleaseNoteID    uuid.UUID
	BugID            uuid.UUID
	ManagerID        uuid.UUID
	OriginalContent  string
	CorrectedContent string
	FeedbackText     *string
	Action           string // "approved_with_correction"
}

// feedbackService implements FeedbackService
type feedbackService struct {
	feedbackRepo repository.FeedbackRepository
	bugRepo      repository.BugRepository
	patternSvc   PatternService
}

// NewFeedbackService creates a new feedback service
func NewFeedbackService(
	feedbackRepo repository.FeedbackRepository,
	bugRepo repository.BugRepository,
	patternSvc PatternService,
) FeedbackService {
	return &feedbackService{
		feedbackRepo: feedbackRepo,
		bugRepo:      bugRepo,
		patternSvc:   patternSvc,
	}
}

// CaptureFeedback captures manager feedback and triggers pattern extraction
func (s *feedbackService) CaptureFeedback(ctx context.Context, req *CaptureFeedbackRequest) (*models.Feedback, error) {
	logger.Info().
		Str("release_note_id", req.ReleaseNoteID.String()).
		Str("manager_id", req.ManagerID.String()).
		Msg("Capturing manager feedback")

	// Get bug for context
	bug, err := s.bugRepo.FindByID(req.BugID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to find bug")
		return nil, fmt.Errorf("failed to find bug: %w", err)
	}

	// Extract bug context for similarity matching
	bugContext := extractBugContext(bug)
	bugContextJSON, _ := json.Marshal(bugContext)

	// Create feedback record
	feedback := &models.Feedback{
		ReleaseNoteID:     req.ReleaseNoteID,
		BugID:             req.BugID,
		ManagerID:         req.ManagerID,
		OriginalContent:   req.OriginalContent,
		CorrectedContent:  req.CorrectedContent,
		FeedbackText:      req.FeedbackText,
		Action:            req.Action,
		BugContext:        bugContextJSON,
		PatternsExtracted: false,
		ExtractedPatterns: []byte("{}"),
	}

	// Save feedback
	if err := s.feedbackRepo.Create(feedback); err != nil {
		logger.Error().Err(err).Msg("Failed to create feedback")
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	logger.Info().
		Str("feedback_id", feedback.ID.String()).
		Msg("Feedback captured successfully")

	// Trigger async pattern extraction
	go func() {
		if err := s.patternSvc.ExtractPatternsFromFeedback(context.Background(), feedback.ID); err != nil {
			logger.Error().
				Err(err).
				Str("feedback_id", feedback.ID.String()).
				Msg("Failed to extract patterns from feedback")
		}
	}()

	return feedback, nil
}

// GetFeedback retrieves feedback by ID
func (s *feedbackService) GetFeedback(ctx context.Context, id uuid.UUID) (*models.Feedback, error) {
	return s.feedbackRepo.FindByID(id)
}

// GetFeedbackByReleaseNote retrieves feedback by release note ID
func (s *feedbackService) GetFeedbackByReleaseNote(ctx context.Context, releaseNoteID uuid.UUID) (*models.Feedback, error) {
	return s.feedbackRepo.FindByReleaseNoteID(releaseNoteID)
}

// GetManagerFeedback retrieves all feedback by a manager
func (s *feedbackService) GetManagerFeedback(ctx context.Context, managerID uuid.UUID, page, limit int) ([]*models.Feedback, int64, error) {
	pagination := &repository.Pagination{
		Page:  page,
		Limit: limit,
	}
	return s.feedbackRepo.FindByManagerID(managerID, pagination)
}

// UpdateEffectivenessScore updates the effectiveness score for feedback
func (s *feedbackService) UpdateEffectivenessScore(ctx context.Context, feedbackID uuid.UUID, score float64) error {
	feedback, err := s.feedbackRepo.FindByID(feedbackID)
	if err != nil {
		return err
	}

	feedback.EffectivenessScore = &score
	return s.feedbackRepo.Update(feedback)
}

// IncrementUsageCount increments the times_used_as_example counter
func (s *feedbackService) IncrementUsageCount(ctx context.Context, feedbackID uuid.UUID) error {
	feedback, err := s.feedbackRepo.FindByID(feedbackID)
	if err != nil {
		return err
	}

	feedback.TimesUsedAsExample++
	return s.feedbackRepo.Update(feedback)
}

// extractBugContext extracts relevant context from bug for similarity matching
func extractBugContext(bug *models.Bug) map[string]interface{} {
	context := make(map[string]interface{})

	// Basic fields
	context["component"] = bug.Component
	context["severity"] = bug.Severity
	context["release"] = bug.Release

	// Check for CVE in title or description
	hasCVE := false
	cveNumber := ""
	if len(bug.Title) > 0 && containsCVE(bug.Title) {
		hasCVE = true
		cveNumber = extractCVENumber(bug.Title)
	}
	context["has_cve"] = hasCVE
	if cveNumber != "" {
		context["cve_number"] = cveNumber
	}

	// Extract keywords from title
	if len(bug.Title) > 0 {
		keywords := extractKeywords(bug.Title)
		context["title_keywords"] = keywords
	}

	// Determine bug type based on keywords
	bugType := determineBugType(bug)
	context["bug_type"] = bugType

	return context
}

// Helper functions for bug context extraction

func containsCVE(text string) bool {
	// Simple check for CVE pattern
	return len(text) > 0 && (contains(text, "CVE-") ||
		contains(text, "cve-") ||
		contains(text, "vulnerability") ||
		contains(text, "security"))
}

func extractCVENumber(text string) string {
	// Simple CVE extraction - can be enhanced with regex
	// Format: CVE-YYYY-NNNNN
	if idx := indexOf(text, "CVE-"); idx != -1 && len(text) > idx+13 {
		return text[idx : idx+13]
	}
	return ""
}

func extractKeywords(title string) []string {
	// Simple keyword extraction - split by spaces and filter
	words := splitWords(title)
	keywords := []string{}

	// Filter out common words and keep meaningful ones
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
	}

	for _, word := range words {
		lower := toLower(word)
		if len(lower) > 2 && !stopWords[lower] {
			keywords = append(keywords, lower)
		}
	}

	return keywords
}

func determineBugType(bug *models.Bug) string {
	// Determine bug type based on title and description
	if len(bug.Title) == 0 {
		return "general"
	}

	title := toLower(bug.Title)

	if contains(title, "security") || contains(title, "vulnerability") || contains(title, "cve") {
		return "security"
	}
	if contains(title, "crash") || contains(title, "panic") || contains(title, "segfault") {
		return "crash"
	}
	if contains(title, "performance") || contains(title, "slow") || contains(title, "latency") {
		return "performance"
	}
	if contains(title, "memory") || contains(title, "leak") {
		return "memory"
	}

	return "general"
}

// String helper functions
func contains(s, substr string) bool {
	return indexOf(s, substr) != -1
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func splitWords(s string) []string {
	words := []string{}
	word := ""

	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			word += string(c)
		} else if len(word) > 0 {
			words = append(words, word)
			word = ""
		}
	}

	if len(word) > 0 {
		words = append(words, word)
	}

	return words
}
