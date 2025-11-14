package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/external/gemini"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"github.com/omnikam04/release-notes-generator/internal/repository"
	"github.com/rs/zerolog/log"
)

var patternLogger = log.With().Str("service", "pattern").Logger()

// PatternService handles pattern extraction and matching
type PatternService interface {
	// Pattern extraction
	ExtractPatternsFromFeedback(ctx context.Context, feedbackID uuid.UUID) error
	ProcessUnprocessedFeedback(ctx context.Context, limit int) error

	// Pattern matching
	FindMatchingPatterns(ctx context.Context, bugContext map[string]interface{}) ([]*models.Pattern, error)
	GetBestExamplesForBug(ctx context.Context, bug *models.Bug, limit int) ([]*models.Feedback, error)

	// Pattern management
	GetPattern(ctx context.Context, id uuid.UUID) (*models.Pattern, error)
	GetAllPatterns(ctx context.Context, page, limit int) ([]*models.Pattern, int64, error)
	GetTopPatterns(ctx context.Context, limit int) ([]*models.Pattern, error)
	DeactivatePattern(ctx context.Context, id uuid.UUID) error
	MergePatterns(ctx context.Context, sourceID, targetID uuid.UUID) error
}

// PatternExtractionResponse represents AI's pattern extraction output
type PatternExtractionResponse struct {
	Patterns          []ExtractedPattern `json:"patterns"`
	OverallConfidence float64            `json:"overall_confidence"`
}

// ExtractedPattern represents a single extracted pattern
type ExtractedPattern struct {
	PatternName string  `json:"pattern_name"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

// patternService implements PatternService
type patternService struct {
	patternRepo         repository.PatternRepository
	feedbackRepo        repository.FeedbackRepository
	feedbackPatternRepo repository.FeedbackPatternRepository
	geminiClient        *gemini.Client
}

// NewPatternService creates a new pattern service
func NewPatternService(
	patternRepo repository.PatternRepository,
	feedbackRepo repository.FeedbackRepository,
	feedbackPatternRepo repository.FeedbackPatternRepository,
	geminiClient *gemini.Client,
) PatternService {
	return &patternService{
		patternRepo:         patternRepo,
		feedbackRepo:        feedbackRepo,
		feedbackPatternRepo: feedbackPatternRepo,
		geminiClient:        geminiClient,
	}
}

// ExtractPatternsFromFeedback uses AI to extract patterns from manager feedback
func (s *patternService) ExtractPatternsFromFeedback(ctx context.Context, feedbackID uuid.UUID) error {
	patternLogger.Info().Str("feedback_id", feedbackID.String()).Msg("Starting pattern extraction")

	// Get feedback
	feedback, err := s.feedbackRepo.FindByID(feedbackID)
	if err != nil {
		return fmt.Errorf("failed to find feedback: %w", err)
	}

	// Skip if already processed
	if feedback.PatternsExtracted {
		patternLogger.Info().Str("feedback_id", feedbackID.String()).Msg("Patterns already extracted")
		return nil
	}

	// Build AI prompt for pattern extraction
	prompt := buildPatternExtractionPrompt(feedback)

	// Call Gemini AI
	response, err := s.geminiClient.GenerateContent(ctx, prompt)
	if err != nil {
		errMsg := fmt.Sprintf("AI pattern extraction failed: %v", err)
		feedback.ExtractionError = &errMsg
		feedback.PatternsExtracted = false
		s.feedbackRepo.Update(feedback)
		return fmt.Errorf("failed to extract patterns: %w", err)
	}

	// Parse AI response
	var extractionResult PatternExtractionResponse
	if err := json.Unmarshal([]byte(response), &extractionResult); err != nil {
		errMsg := fmt.Sprintf("Failed to parse AI response: %v", err)
		feedback.ExtractionError = &errMsg
		feedback.PatternsExtracted = false
		s.feedbackRepo.Update(feedback)
		return fmt.Errorf("failed to parse pattern extraction response: %w", err)
	}

	patternLogger.Info().
		Str("feedback_id", feedbackID.String()).
		Int("patterns_found", len(extractionResult.Patterns)).
		Float64("confidence", extractionResult.OverallConfidence).
		Msg("Patterns extracted successfully")

	// Store extracted patterns in feedback
	extractedJSON, _ := json.Marshal(extractionResult)
	feedback.ExtractedPatterns = extractedJSON
	feedback.OverallConfidence = extractionResult.OverallConfidence
	feedback.PatternsExtracted = true
	feedback.ExtractionError = nil

	if err := s.feedbackRepo.Update(feedback); err != nil {
		return fmt.Errorf("failed to update feedback: %w", err)
	}

	// Process each extracted pattern
	for _, extractedPattern := range extractionResult.Patterns {
		if err := s.processExtractedPattern(ctx, feedback, &extractedPattern); err != nil {
			patternLogger.Error().
				Err(err).
				Str("pattern_name", extractedPattern.PatternName).
				Msg("Failed to process extracted pattern")
			// Continue with other patterns
		}
	}

	return nil
}

// processExtractedPattern creates or updates a pattern and links it to feedback
func (s *patternService) processExtractedPattern(ctx context.Context, feedback *models.Feedback, extracted *ExtractedPattern) error {
	// Check if pattern already exists
	pattern, err := s.patternRepo.FindByName(extracted.PatternName)
	if err != nil {
		// Pattern doesn't exist - create new one
		pattern = &models.Pattern{
			Name:            extracted.PatternName,
			Category:        extracted.Category,
			Description:     extracted.Description,
			OccurrenceCount: 1,
			AvgConfidence:   extracted.Confidence,
			Priority:        calculatePriority(extracted.Category),
			IsActive:        true,
		}

		// Set applicable_when based on bug context
		pattern.ApplicableWhen = feedback.BugContext

		if err := s.patternRepo.Create(pattern); err != nil {
			return fmt.Errorf("failed to create pattern: %w", err)
		}

		patternLogger.Info().
			Str("pattern_name", pattern.Name).
			Str("category", pattern.Category).
			Msg("New pattern created")
	} else {
		// Pattern exists - update statistics
		if err := s.patternRepo.UpdateStatistics(pattern.ID, extracted.Confidence, true); err != nil {
			return fmt.Errorf("failed to update pattern statistics: %w", err)
		}
	}

	// Create feedback-pattern link
	feedbackPattern := &models.FeedbackPattern{
		FeedbackID:  feedback.ID,
		PatternID:   pattern.ID,
		Confidence:  extracted.Confidence,
		Description: extracted.Description,
	}

	if err := s.feedbackPatternRepo.Create(feedbackPattern); err != nil {
		return fmt.Errorf("failed to create feedback-pattern link: %w", err)
	}

	return nil
}

// ProcessUnprocessedFeedback processes all feedback that hasn't had patterns extracted
func (s *patternService) ProcessUnprocessedFeedback(ctx context.Context, limit int) error {
	feedbacks, err := s.feedbackRepo.FindUnprocessedFeedback(limit)
	if err != nil {
		return err
	}

	patternLogger.Info().Int("count", len(feedbacks)).Msg("Processing unprocessed feedback")

	for _, feedback := range feedbacks {
		if err := s.ExtractPatternsFromFeedback(ctx, feedback.ID); err != nil {
			patternLogger.Error().
				Err(err).
				Str("feedback_id", feedback.ID.String()).
				Msg("Failed to extract patterns")
			// Continue with other feedback
		}
	}

	return nil
}

// FindMatchingPatterns finds patterns that apply to the given bug context
func (s *patternService) FindMatchingPatterns(ctx context.Context, bugContext map[string]interface{}) ([]*models.Pattern, error) {
	// Get all active patterns
	allPatterns, err := s.patternRepo.FindActivePatterns()
	if err != nil {
		return nil, err
	}

	// Filter patterns that match bug context
	matchingPatterns := []*models.Pattern{}
	for _, pattern := range allPatterns {
		if patternMatchesBugContext(pattern, bugContext) {
			matchingPatterns = append(matchingPatterns, pattern)
		}
	}

	return matchingPatterns, nil
}

// GetBestExamplesForBug finds the best feedback examples for a given bug
func (s *patternService) GetBestExamplesForBug(ctx context.Context, bug *models.Bug, limit int) ([]*models.Feedback, error) {
	// Extract bug context
	bugContext := extractBugContext(bug)

	// Find similar feedback
	examples, err := s.feedbackRepo.FindSimilarFeedback(bugContext, limit)
	if err != nil {
		return nil, err
	}

	// If not enough similar examples, get most effective ones
	if len(examples) < limit {
		additional, err := s.feedbackRepo.FindMostEffectiveFeedback(limit - len(examples))
		if err == nil {
			examples = append(examples, additional...)
		}
	}

	return examples, nil
}

// GetPattern retrieves a pattern by ID
func (s *patternService) GetPattern(ctx context.Context, id uuid.UUID) (*models.Pattern, error) {
	return s.patternRepo.FindByID(id)
}

// GetAllPatterns retrieves all patterns with pagination
func (s *patternService) GetAllPatterns(ctx context.Context, page, limit int) ([]*models.Pattern, int64, error) {
	pagination := &repository.Pagination{
		Page:  page,
		Limit: limit,
	}
	return s.patternRepo.ListAll(pagination)
}

// GetTopPatterns retrieves the most successful patterns
func (s *patternService) GetTopPatterns(ctx context.Context, limit int) ([]*models.Pattern, error) {
	return s.patternRepo.FindTopPatterns(limit)
}

// DeactivatePattern marks a pattern as inactive
func (s *patternService) DeactivatePattern(ctx context.Context, id uuid.UUID) error {
	return s.patternRepo.DeactivatePattern(id)
}

// MergePatterns merges two similar patterns
func (s *patternService) MergePatterns(ctx context.Context, sourceID, targetID uuid.UUID) error {
	return s.patternRepo.MergePatterns(sourceID, targetID)
}

// Helper functions

func buildPatternExtractionPrompt(feedback *models.Feedback) string {
	var bugContextStr string
	if len(feedback.BugContext) > 0 {
		bugContextStr = string(feedback.BugContext)
	} else {
		bugContextStr = "{}"
	}

	feedbackText := ""
	if feedback.FeedbackText != nil {
		feedbackText = *feedback.FeedbackText
	}

	prompt := fmt.Sprintf(`You are a pattern extraction expert for release note quality improvement.

Analyze the differences between the AI-generated and manager-corrected release notes.

ORIGINAL (AI-generated):
%s

CORRECTED (Manager's version):
%s

MANAGER FEEDBACK:
%s

BUG CONTEXT:
%s

Extract specific patterns that explain what went wrong and how to improve.

PATTERN CATEGORIES:
- clarity: Issues with clarity, jargon, technical language
- style: Issues with writing style, tone, voice
- content: Missing or incorrect content
- structure: Issues with sentence structure, length
- consistency: Inconsistency with standards or conventions

OUTPUT (JSON format only, no markdown):
{
  "patterns": [
    {
      "pattern_name": "snake_case_name",
      "confidence": 0.95,
      "description": "Brief description of the pattern",
      "category": "clarity"
    }
  ],
  "overall_confidence": 0.92
}

EXAMPLES OF GOOD PATTERN NAMES:
- "too_technical_jargon"
- "abbreviation_expansion"
- "verb_consistency"
- "missing_device_specificity"
- "passive_voice_usage"
- "exceeds_length_limit"
- "missing_cve_reference"
- "customer_facing_language"

Extract 1-5 patterns. Focus on the most significant differences.
Return ONLY the JSON object, no additional text.`,
		feedback.OriginalContent,
		feedback.CorrectedContent,
		feedbackText,
		bugContextStr,
	)

	return prompt
}

func calculatePriority(category string) int {
	// Assign priority based on category
	priorities := map[string]int{
		"content":     100, // Highest priority - missing/incorrect content
		"clarity":     80,  // High priority - clarity issues
		"consistency": 60,  // Medium priority - consistency
		"structure":   40,  // Lower priority - structure
		"style":       20,  // Lowest priority - style preferences
	}

	if priority, ok := priorities[category]; ok {
		return priority
	}
	return 50 // Default priority
}

func patternMatchesBugContext(pattern *models.Pattern, bugContext map[string]interface{}) bool {
	// If pattern has no applicability criteria, it applies to all bugs
	if len(pattern.ApplicableWhen) == 0 {
		return true
	}

	// Parse applicable_when JSON
	var applicableWhen map[string]interface{}
	if err := json.Unmarshal(pattern.ApplicableWhen, &applicableWhen); err != nil {
		return false
	}

	// Check if all conditions in applicable_when are met by bugContext
	for key, value := range applicableWhen {
		bugValue, exists := bugContext[key]
		if !exists {
			return false
		}

		// Simple equality check - can be enhanced for array/complex matching
		if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", bugValue) {
			return false
		}
	}

	return true
}
