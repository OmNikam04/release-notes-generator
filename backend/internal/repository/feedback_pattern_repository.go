package repository

import (
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

// FeedbackPatternRepository defines the interface for feedback-pattern junction operations
type FeedbackPatternRepository interface {
	Create(feedbackPattern *models.FeedbackPattern) error
	CreateBatch(feedbackPatterns []*models.FeedbackPattern) error
	FindByID(id uuid.UUID) (*models.FeedbackPattern, error)
	FindByFeedbackID(feedbackID uuid.UUID) ([]*models.FeedbackPattern, error)
	FindByPatternID(patternID uuid.UUID) ([]*models.FeedbackPattern, error)
	Update(feedbackPattern *models.FeedbackPattern) error
	Delete(id uuid.UUID) error
	
	// Effectiveness tracking
	MarkAsHelpful(id uuid.UUID, wasHelpful bool) error
	FindMostHelpfulForPattern(patternID uuid.UUID, limit int) ([]*models.FeedbackPattern, error)
}

// feedbackPatternRepository is the concrete implementation
type feedbackPatternRepository struct {
	db *gorm.DB
}

// NewFeedbackPatternRepository creates a new feedback-pattern repository instance
func NewFeedbackPatternRepository(db *gorm.DB) FeedbackPatternRepository {
	return &feedbackPatternRepository{db: db}
}

// Create creates a new feedback-pattern link
func (r *feedbackPatternRepository) Create(feedbackPattern *models.FeedbackPattern) error {
	return r.db.Create(feedbackPattern).Error
}

// CreateBatch creates multiple feedback-pattern links in a transaction
func (r *feedbackPatternRepository) CreateBatch(feedbackPatterns []*models.FeedbackPattern) error {
	if len(feedbackPatterns) == 0 {
		return nil
	}
	return r.db.Create(&feedbackPatterns).Error
}

// FindByID finds a feedback-pattern link by its ID
func (r *feedbackPatternRepository) FindByID(id uuid.UUID) (*models.FeedbackPattern, error) {
	var fp models.FeedbackPattern
	err := r.db.
		Preload("Feedback").
		Preload("Pattern").
		First(&fp, "id = ?", id).Error
	return &fp, err
}

// FindByFeedbackID finds all patterns linked to a feedback
func (r *feedbackPatternRepository) FindByFeedbackID(feedbackID uuid.UUID) ([]*models.FeedbackPattern, error) {
	var fps []*models.FeedbackPattern
	err := r.db.
		Where("feedback_id = ?", feedbackID).
		Preload("Pattern").
		Order("confidence DESC").
		Find(&fps).Error
	return fps, err
}

// FindByPatternID finds all feedback linked to a pattern
func (r *feedbackPatternRepository) FindByPatternID(patternID uuid.UUID) ([]*models.FeedbackPattern, error) {
	var fps []*models.FeedbackPattern
	err := r.db.
		Where("pattern_id = ?", patternID).
		Preload("Feedback").
		Order("confidence DESC").
		Find(&fps).Error
	return fps, err
}

// Update updates an existing feedback-pattern link
func (r *feedbackPatternRepository) Update(feedbackPattern *models.FeedbackPattern) error {
	return r.db.Save(feedbackPattern).Error
}

// Delete deletes a feedback-pattern link
func (r *feedbackPatternRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.FeedbackPattern{}, "id = ?", id).Error
}

// MarkAsHelpful marks a feedback-pattern link as helpful or not
func (r *feedbackPatternRepository) MarkAsHelpful(id uuid.UUID, wasHelpful bool) error {
	return r.db.Model(&models.FeedbackPattern{}).
		Where("id = ?", id).
		Update("was_helpful", wasHelpful).
		Error
}

// FindMostHelpfulForPattern finds the most helpful feedback examples for a pattern
func (r *feedbackPatternRepository) FindMostHelpfulForPattern(patternID uuid.UUID, limit int) ([]*models.FeedbackPattern, error) {
	var fps []*models.FeedbackPattern
	err := r.db.
		Where("pattern_id = ? AND was_helpful = ?", patternID, true).
		Preload("Feedback").
		Preload("Feedback.Bug").
		Preload("Feedback.ReleaseNote").
		Order("confidence DESC").
		Limit(limit).
		Find(&fps).Error
	return fps, err
}

