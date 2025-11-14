package repository

import (
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

// PatternRepository defines the interface for pattern data operations
type PatternRepository interface {
	Create(pattern *models.Pattern) error
	FindByID(id uuid.UUID) (*models.Pattern, error)
	FindByName(name string) (*models.Pattern, error)
	FindByCategory(category string) ([]*models.Pattern, error)
	Update(pattern *models.Pattern) error
	Delete(id uuid.UUID) error

	// Pattern matching queries
	FindActivePatterns() ([]*models.Pattern, error)
	FindMatchingPatterns(bugContext map[string]interface{}) ([]*models.Pattern, error)
	FindTopPatterns(limit int) ([]*models.Pattern, error)

	// Pattern statistics
	IncrementOccurrence(id uuid.UUID) error
	UpdateStatistics(id uuid.UUID, confidence float64, wasSuccessful bool) error

	// Pattern management
	ListAll(pagination *Pagination) ([]*models.Pattern, int64, error)
	DeactivatePattern(id uuid.UUID) error
	MergePatterns(sourceID, targetID uuid.UUID) error
}

// patternRepository is the concrete implementation
type patternRepository struct {
	db *gorm.DB
}

// NewPatternRepository creates a new pattern repository instance
func NewPatternRepository(db *gorm.DB) PatternRepository {
	return &patternRepository{db: db}
}

// Create creates a new pattern
func (r *patternRepository) Create(pattern *models.Pattern) error {
	return r.db.Create(pattern).Error
}

// FindByID finds a pattern by its ID
func (r *patternRepository) FindByID(id uuid.UUID) (*models.Pattern, error) {
	var pattern models.Pattern
	err := r.db.
		Preload("FeedbackPatterns").
		First(&pattern, "id = ?", id).Error
	return &pattern, err
}

// FindByName finds a pattern by its name (unique)
func (r *patternRepository) FindByName(name string) (*models.Pattern, error) {
	var pattern models.Pattern
	err := r.db.First(&pattern, "name = ?", name).Error
	return &pattern, err
}

// FindByCategory finds all patterns in a category
func (r *patternRepository) FindByCategory(category string) ([]*models.Pattern, error) {
	var patterns []*models.Pattern
	err := r.db.
		Where("category = ? AND is_active = ?", category, true).
		Order("priority DESC, occurrence_count DESC").
		Find(&patterns).Error
	return patterns, err
}

// Update updates an existing pattern
func (r *patternRepository) Update(pattern *models.Pattern) error {
	return r.db.Save(pattern).Error
}

// Delete soft deletes a pattern
func (r *patternRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Pattern{}, "id = ?", id).Error
}

// FindActivePatterns finds all active patterns
func (r *patternRepository) FindActivePatterns() ([]*models.Pattern, error) {
	var patterns []*models.Pattern
	err := r.db.
		Where("is_active = ?", true).
		Order("priority DESC, success_rate DESC").
		Find(&patterns).Error
	return patterns, err
}

// FindMatchingPatterns finds patterns that match the given bug context
func (r *patternRepository) FindMatchingPatterns(bugContext map[string]interface{}) ([]*models.Pattern, error) {
	var patterns []*models.Pattern

	// Start with all active patterns
	query := r.db.Where("is_active = ?", true)

	// For now, simple implementation - can be enhanced with JSONB queries
	// Match patterns where applicable_when is empty (applies to all) or matches bug context

	// PostgreSQL JSONB containment check
	// Pattern applies if its applicable_when is a subset of bug_context
	// For simplicity, we'll fetch all active patterns and filter in service layer
	// In production, use: query = query.Where("applicable_when <@ ?", bugContext)

	err := query.
		Order("priority DESC, success_rate DESC").
		Find(&patterns).Error

	return patterns, err
}

// FindTopPatterns finds the most successful patterns
func (r *patternRepository) FindTopPatterns(limit int) ([]*models.Pattern, error) {
	var patterns []*models.Pattern
	err := r.db.
		Where("is_active = ? AND occurrence_count > ?", true, 0).
		Order("success_rate DESC, occurrence_count DESC").
		Limit(limit).
		Find(&patterns).Error
	return patterns, err
}

// IncrementOccurrence increments the occurrence count for a pattern
func (r *patternRepository) IncrementOccurrence(id uuid.UUID) error {
	return r.db.Model(&models.Pattern{}).
		Where("id = ?", id).
		UpdateColumn("occurrence_count", gorm.Expr("occurrence_count + 1")).
		Error
}

// UpdateStatistics updates pattern statistics based on new feedback
func (r *patternRepository) UpdateStatistics(id uuid.UUID, confidence float64, wasSuccessful bool) error {
	var pattern models.Pattern
	if err := r.db.First(&pattern, "id = ?", id).Error; err != nil {
		return err
	}

	// Update average confidence (running average)
	totalConfidence := pattern.AvgConfidence * float64(pattern.OccurrenceCount)
	newCount := pattern.OccurrenceCount + 1
	pattern.AvgConfidence = (totalConfidence + confidence) / float64(newCount)

	// Update success rate if we have success/failure data
	if wasSuccessful {
		// Simple increment - in production, track successes and failures separately
		pattern.SuccessRate = (pattern.SuccessRate*float64(pattern.OccurrenceCount) + 1.0) / float64(newCount)
	}

	pattern.OccurrenceCount = newCount

	return r.db.Save(&pattern).Error
}

// ListAll lists all patterns with pagination
func (r *patternRepository) ListAll(pagination *Pagination) ([]*models.Pattern, int64, error) {
	var patterns []*models.Pattern
	var total int64

	query := r.db.Model(&models.Pattern{})

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if pagination != nil {
		query = r.applyPagination(query, pagination)
	}

	err := query.
		Order("is_active DESC, priority DESC, occurrence_count DESC").
		Find(&patterns).Error

	return patterns, total, err
}

// DeactivatePattern marks a pattern as inactive
func (r *patternRepository) DeactivatePattern(id uuid.UUID) error {
	return r.db.Model(&models.Pattern{}).
		Where("id = ?", id).
		Update("is_active", false).
		Error
}

// MergePatterns merges source pattern into target pattern
func (r *patternRepository) MergePatterns(sourceID, targetID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update all feedback_patterns to point to target
		if err := tx.Model(&models.FeedbackPattern{}).
			Where("pattern_id = ?", sourceID).
			Update("pattern_id", targetID).Error; err != nil {
			return err
		}

		// Mark source as merged
		if err := tx.Model(&models.Pattern{}).
			Where("id = ?", sourceID).
			Updates(map[string]interface{}{
				"is_active":      false,
				"merged_into_id": targetID,
			}).Error; err != nil {
			return err
		}

		// Recalculate target statistics
		var target models.Pattern
		if err := tx.First(&target, "id = ?", targetID).Error; err != nil {
			return err
		}

		var source models.Pattern
		if err := tx.First(&source, "id = ?", sourceID).Error; err != nil {
			return err
		}

		// Merge occurrence counts
		target.OccurrenceCount += source.OccurrenceCount

		// Recalculate average confidence
		totalConfidence := (target.AvgConfidence * float64(target.OccurrenceCount)) +
			(source.AvgConfidence * float64(source.OccurrenceCount))
		target.AvgConfidence = totalConfidence / float64(target.OccurrenceCount+source.OccurrenceCount)

		return tx.Save(&target).Error
	})
}

// applyPagination applies pagination and sorting to the query
func (r *patternRepository) applyPagination(query *gorm.DB, pagination *Pagination) *gorm.DB {
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
