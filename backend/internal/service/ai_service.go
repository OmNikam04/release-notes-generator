package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/external/gemini"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"github.com/rs/zerolog/log"
)

// AIService handles AI-powered release note generation
type AIService interface {
	GenerateReleaseNote(ctx context.Context, bug *models.Bug, commits []*bugsby.ParsedCommitInfo) (*AIReleaseNoteResponse, error)
	GenerateReleaseNoteWithPatterns(ctx context.Context, bug *models.Bug, commits []*bugsby.ParsedCommitInfo, patternSvc PatternService) (*AIReleaseNoteResponse, error)
	Close() error
}

// aiService implements AIService
type aiService struct {
	geminiClient *gemini.Client
	model        string
}

// NewAIService creates a new AI service
func NewAIService(ctx context.Context, cfg *gemini.Config) (AIService, error) {
	client, err := gemini.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &aiService{
		geminiClient: client,
		model:        cfg.Model,
	}, nil
}

// Close closes the AI service and releases resources
func (s *aiService) Close() error {
	if s.geminiClient != nil {
		return s.geminiClient.Close()
	}
	return nil
}

// GenerateReleaseNote generates a release note using AI
func (s *aiService) GenerateReleaseNote(
	ctx context.Context,
	bug *models.Bug,
	commits []*bugsby.ParsedCommitInfo,
) (*AIReleaseNoteResponse, error) {
	// Build prompt based on available information
	var prompt string
	if len(commits) > 0 {
		prompt = BuildReleaseNotePrompt(bug, commits)
		log.Info().
			Str("bug_id", bug.BugsbyID).
			Int("commit_count", len(commits)).
			Msg("Generating release note with commit information")
	} else {
		prompt = BuildReleaseNotePromptSimple(bug)
		log.Info().
			Str("bug_id", bug.BugsbyID).
			Msg("Generating release note without commit information")
	}

	// Call Gemini API
	response, err := s.geminiClient.GenerateContent(ctx, prompt)
	if err != nil {
		log.Error().
			Err(err).
			Str("bug_id", bug.BugsbyID).
			Msg("Failed to generate release note with AI")
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse the JSON response from AI
	aiResponse, err := ParseAIResponse(response)
	if err != nil {
		log.Error().
			Err(err).
			Str("bug_id", bug.BugsbyID).
			Msg("Failed to parse AI response")
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	if aiResponse.ReleaseNote == "" {
		return nil, fmt.Errorf("AI returned empty release note")
	}

	// Apply additional confidence adjustments based on context quality
	aiResponse.Confidence = adjustConfidence(aiResponse.Confidence, bug, commits, aiResponse.ReleaseNote)

	log.Info().
		Str("bug_id", bug.BugsbyID).
		Float64("confidence", aiResponse.Confidence).
		Str("reasoning", aiResponse.Reasoning).
		Int("alternatives", len(aiResponse.AlternativeVersions)).
		Msg("Successfully generated release note with AI")

	return aiResponse, nil
}

// GenerateReleaseNoteWithPatterns generates a release note using AI with pattern-aware few-shot learning
func (s *aiService) GenerateReleaseNoteWithPatterns(
	ctx context.Context,
	bug *models.Bug,
	commits []*bugsby.ParsedCommitInfo,
	patternSvc PatternService,
) (*AIReleaseNoteResponse, error) {
	// Get best examples for this bug
	examples, err := patternSvc.GetBestExamplesForBug(ctx, bug, 3)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get pattern examples, falling back to standard generation")
		return s.GenerateReleaseNote(ctx, bug, commits)
	}

	// If no examples found, use standard generation
	if len(examples) == 0 {
		log.Info().Msg("No pattern examples found, using standard generation")
		return s.GenerateReleaseNote(ctx, bug, commits)
	}

	// Build enhanced prompt with few-shot examples
	var prompt string
	if len(commits) > 0 {
		prompt = BuildReleaseNotePromptWithPatterns(bug, commits, examples)
		log.Info().
			Str("bug_id", bug.BugsbyID).
			Int("commit_count", len(commits)).
			Int("example_count", len(examples)).
			Msg("Generating release note with commit information and pattern examples")
	} else {
		prompt = BuildReleaseNotePromptWithPatternsNoCommits(bug, examples)
		log.Info().
			Str("bug_id", bug.BugsbyID).
			Int("example_count", len(examples)).
			Msg("Generating release note without commits but with pattern examples")
	}

	// Call Gemini AI
	responseText, err := s.geminiClient.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate release note: %w", err)
	}

	// Parse response
	aiResponse, err := parseAIResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Adjust confidence based on context quality
	aiResponse.Confidence = adjustConfidence(aiResponse.Confidence, bug, commits, aiResponse.ReleaseNote)

	log.Info().
		Str("bug_id", bug.BugsbyID).
		Float64("confidence", aiResponse.Confidence).
		Int("examples_used", len(examples)).
		Msg("Release note generated successfully with patterns")

	return aiResponse, nil
}

// adjustConfidence adjusts the AI's confidence score based on context quality
func adjustConfidence(aiConfidence float64, bug *models.Bug, commits []*bugsby.ParsedCommitInfo, content string) float64 {
	confidence := aiConfidence

	// Small boost if we have commits (AI might not account for this)
	if len(commits) > 0 && confidence < 0.9 {
		confidence += 0.05
	}

	// Small boost if bug has detailed description
	if bug.Description != nil && len(*bug.Description) > 100 && confidence < 0.9 {
		confidence += 0.05
	}

	// Small boost if content is well-formed
	if isWellFormedReleaseNote(content) && confidence < 0.9 {
		confidence += 0.05
	}

	// Cap confidence at 0.95 (never 100% certain)
	if confidence > 0.95 {
		confidence = 0.95
	}

	// Ensure minimum confidence
	if confidence < 0.3 {
		confidence = 0.3
	}

	return confidence
}

// isWellFormedReleaseNote checks if the release note is well-formed
func isWellFormedReleaseNote(content string) bool {
	// Check minimum length
	if len(content) < 50 {
		return false
	}

	// Check maximum length (should be concise)
	if len(content) > 1000 {
		return false
	}

	// Check if it starts with a capital letter
	if len(content) > 0 && !isUpperCase(rune(content[0])) {
		return false
	}

	// Check if it ends with proper punctuation
	lastChar := rune(content[len(content)-1])
	if lastChar != '.' && lastChar != '!' && lastChar != '?' {
		return false
	}

	// Check if it contains common release note keywords
	keywords := []string{"fixed", "resolved", "corrected", "addressed", "improved", "updated"}
	contentLower := strings.ToLower(content)
	hasKeyword := false
	for _, keyword := range keywords {
		if strings.Contains(contentLower, keyword) {
			hasKeyword = true
			break
		}
	}

	return hasKeyword
}

// isUpperCase checks if a rune is uppercase
func isUpperCase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}
