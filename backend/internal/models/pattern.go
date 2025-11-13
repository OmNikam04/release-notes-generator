package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Pattern represents an identified pattern for smart example selection
type Pattern struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Pattern Identity
	Name        string `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"` // e.g., "missing_cve_reference"
	Category    string `json:"category" gorm:"type:varchar(50);not null;index"`    // "clarity", "style", "content", "structure", "consistency"
	Description string `json:"description" gorm:"type:text;not null"`              // Human-readable description

	// Pattern Matching Criteria
	// This defines WHEN this pattern applies (bug characteristics)
	ApplicableWhen datatypes.JSON `json:"applicable_when" gorm:"type:jsonb;not null;default:'{}'"`
	// Example: {
	//   "bug_type": ["security"],
	//   "severity": ["high", "critical"],
	//   "has_cve": true
	// }

	// Pattern Examples (Best examples of this pattern)
	ExampleFeedbackIDs pq.StringArray `json:"example_feedback_ids" gorm:"type:uuid[]"` // Array of feedback IDs that exemplify this pattern

	// Pattern Statistics
	OccurrenceCount int     `json:"occurrence_count" gorm:"default:0"`                      // How many times this pattern was extracted
	SuccessRate     float64 `json:"success_rate" gorm:"type:decimal(3,2);default:0"`        // Success rate when applied
	AvgConfidence   float64 `json:"avg_confidence" gorm:"type:decimal(3,2);default:0"`      // Average confidence from AI extractions

	// Pattern Priority & Status
	Priority int  `json:"priority" gorm:"default:0;index"`       // Higher = more important
	IsActive bool `json:"is_active" gorm:"default:true;index"`   // Whether to use in matching

	// Pattern Relationships (for merging similar patterns)
	SimilarPatternIDs pq.StringArray `json:"similar_pattern_ids" gorm:"type:uuid[]"` // Related patterns
	MergedIntoID      *uuid.UUID     `json:"merged_into_id" gorm:"type:uuid"`        // If merged into another pattern

	// Relationships
	FeedbackPatterns []FeedbackPattern `json:"feedback_patterns,omitempty" gorm:"foreignKey:PatternID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID
func (p *Pattern) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for Pattern model
func (Pattern) TableName() string {
	return "patterns"
}

