package repository

import (
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

// FeedbackRepository defines the interface for feedback data operations
type FeedbackRepository interface {
	Create(feedback *models.Feedback) error
	FindByID(id uuid.UUID) (*models.Feedback, error)
	FindByReleaseNoteID(releaseNoteID uuid.UUID) (*models.Feedback, error)
	FindByManagerID(managerID uuid.UUID, pagination *Pagination) ([]*models.Feedback, int64, error)
	Update(feedback *models.Feedback) error
	Delete(id uuid.UUID) error

	// Pattern extraction queries
	FindUnprocessedFeedback(limit int) ([]*models.Feedback, error)
	FindByPatternID(patternID uuid.UUID, limit int) ([]*models.Feedback, error)

	// Smart example selection
	FindSimilarFeedback(bugContext map[string]interface{}, limit int) ([]*models.Feedback, error)
	FindMostEffectiveFeedback(limit int) ([]*models.Feedback, error)
}

// feedbackRepository is the concrete implementation
type feedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository creates a new feedback repository instance
func NewFeedbackRepository(db *gorm.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

// Create creates a new feedback record
func (r *feedbackRepository) Create(feedback *models.Feedback) error {
	return r.db.Create(feedback).Error
}

// FindByID finds a feedback by its ID
func (r *feedbackRepository) FindByID(id uuid.UUID) (*models.Feedback, error) {
	var feedback models.Feedback
	err := r.db.
		Preload("ReleaseNote").
		Preload("Bug").
		Preload("Manager").
		Preload("FeedbackPatterns").
		First(&feedback, "id = ?", id).Error
	return &feedback, err
}

// FindByReleaseNoteID finds feedback by release note ID
func (r *feedbackRepository) FindByReleaseNoteID(releaseNoteID uuid.UUID) (*models.Feedback, error) {
	var feedback models.Feedback
	err := r.db.
		Preload("ReleaseNote").
		Preload("Bug").
		Preload("Manager").
		Preload("FeedbackPatterns").
		First(&feedback, "release_note_id = ?", releaseNoteID).Error
	return &feedback, err
}

// FindByManagerID finds all feedback by a specific manager
func (r *feedbackRepository) FindByManagerID(managerID uuid.UUID, pagination *Pagination) ([]*models.Feedback, int64, error) {
	var feedbacks []*models.Feedback
	var total int64

	query := r.db.Model(&models.Feedback{}).Where("manager_id = ?", managerID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	// Execute query
	err := query.
		Preload("ReleaseNote").
		Preload("Bug").
		Preload("FeedbackPatterns").
		Find(&feedbacks).Error

	return feedbacks, total, err
}

// Update updates an existing feedback record
func (r *feedbackRepository) Update(feedback *models.Feedback) error {
	return r.db.Save(feedback).Error
}

// Delete soft deletes a feedback record
func (r *feedbackRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Feedback{}, "id = ?", id).Error
}

// FindUnprocessedFeedback finds feedback that hasn't had patterns extracted yet
func (r *feedbackRepository) FindUnprocessedFeedback(limit int) ([]*models.Feedback, error) {
	var feedbacks []*models.Feedback
	err := r.db.
		Where("patterns_extracted = ?", false).
		Preload("ReleaseNote").
		Preload("Bug").
		Limit(limit).
		Find(&feedbacks).Error
	return feedbacks, err
}

// FindByPatternID finds all feedback associated with a specific pattern
func (r *feedbackRepository) FindByPatternID(patternID uuid.UUID, limit int) ([]*models.Feedback, error) {
	var feedbacks []*models.Feedback
	err := r.db.
		Joins("JOIN feedback_patterns ON feedback_patterns.feedback_id = feedbacks.id").
		Where("feedback_patterns.pattern_id = ?", patternID).
		Preload("ReleaseNote").
		Preload("Bug").
		Preload("FeedbackPatterns").
		Order("feedbacks.effectiveness_score DESC NULLS LAST").
		Limit(limit).
		Find(&feedbacks).Error
	return feedbacks, err
}

// FindSimilarFeedback finds feedback with similar bug context (for smart example selection)
func (r *feedbackRepository) FindSimilarFeedback(bugContext map[string]interface{}, limit int) ([]*models.Feedback, error) {
	var feedbacks []*models.Feedback

	// For now, simple implementation - can be enhanced with vector similarity later
	// Match on component, severity, has_cve, etc.
	query := r.db.Model(&models.Feedback{}).
		Where("patterns_extracted = ?", true).
		Where("effectiveness_score IS NOT NULL")

	// Add JSON containment checks if bug context has specific fields
	// This is PostgreSQL-specific JSONB query
	if component, ok := bugContext["component"].(string); ok && component != "" {
		query = query.Where("bug_context->>'component' = ?", component)
	}

	err := query.
		Preload("ReleaseNote").
		Preload("Bug").
		Preload("FeedbackPatterns.Pattern").
		Order("effectiveness_score DESC, times_used_as_example ASC").
		Limit(limit).
		Find(&feedbacks).Error

	return feedbacks, err
}

// FindMostEffectiveFeedback finds the most effective feedback examples
func (r *feedbackRepository) FindMostEffectiveFeedback(limit int) ([]*models.Feedback, error) {
	var feedbacks []*models.Feedback
	err := r.db.
		Where("patterns_extracted = ?", true).
		Where("effectiveness_score IS NOT NULL").
		Preload("ReleaseNote").
		Preload("Bug").
		Preload("FeedbackPatterns.Pattern").
		Order("effectiveness_score DESC, overall_confidence DESC").
		Limit(limit).
		Find(&feedbacks).Error
	return feedbacks, err
}

// applyPagination applies pagination and sorting to the query
func (r *feedbackRepository) applyPagination(query *gorm.DB, pagination *Pagination) *gorm.DB {
	// Set defaults
	page := pagination.Page
	if page < 1 {
		page = 1
	}

	limit := pagination.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset := (page - 1) * limit

	// Apply sorting
	sortBy := pagination.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := pagination.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Apply pagination
	return query.Offset(offset).Limit(limit)
}
