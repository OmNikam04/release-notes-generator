package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FeedbackPattern links feedback to patterns with confidence scores (junction table)
type FeedbackPattern struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	FeedbackID uuid.UUID `json:"feedback_id" gorm:"type:uuid;not null;index"` // Foreign key to feedbacks
	PatternID  uuid.UUID `json:"pattern_id" gorm:"type:uuid;not null;index"`  // Foreign key to patterns

	// Pattern Details (from AI extraction)
	Confidence  float64 `json:"confidence" gorm:"type:decimal(3,2);not null"` // 0.0-1.0
	Description string  `json:"description" gorm:"type:text"`                 // Specific description for this instance

	// Effectiveness Tracking
	WasHelpful *bool `json:"was_helpful"` // Did this pattern help improve future generations? (nullable)

	// Relationships
	Feedback *Feedback `json:"feedback,omitempty" gorm:"foreignKey:FeedbackID;constraint:OnDelete:CASCADE"`
	Pattern  *Pattern  `json:"pattern,omitempty" gorm:"foreignKey:PatternID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID
func (fp *FeedbackPattern) BeforeCreate(tx *gorm.DB) error {
	if fp.ID == uuid.Nil {
		fp.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for FeedbackPattern model
func (FeedbackPattern) TableName() string {
	return "feedback_patterns"
}

