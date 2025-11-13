package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReleaseNote represents a release note for a bug (AI-generated or manually written)
type ReleaseNote struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	BugID uuid.UUID `json:"bug_id" gorm:"type:uuid;uniqueIndex;not null"` // Foreign key to bugs table (one note per bug)

	// Content
	Content string `json:"content" gorm:"type:text;not null"` // The actual release note text
	Version int    `json:"version" gorm:"default:1"`          // Version number (for tracking edits)

	// Generation Info
	GeneratedBy  string   `json:"generated_by" gorm:"type:varchar(20);not null"` // "ai" or "manual"
	AIModel      *string  `json:"ai_model" gorm:"type:varchar(50)"`              // AI model used (e.g., "gpt-4"), nullable
	AIConfidence *float64 `json:"ai_confidence" gorm:"type:decimal(3,2)"`        // AI confidence score (0.0-1.0), nullable

	// Approval Tracking
	Status string `json:"status" gorm:"type:varchar(50);not null;index;default:'draft'"` // "draft", "ai_generated", "dev_approved", "mgr_approved", "rejected"

	// User Actions
	CreatedByID     *uuid.UUID `json:"created_by_id" gorm:"type:uuid;index"`     // User who created (NULL for AI), foreign key
	ApprovedByDevID *uuid.UUID `json:"approved_by_dev_id" gorm:"type:uuid"`      // Developer who approved, foreign key, nullable
	ApprovedByMgrID *uuid.UUID `json:"approved_by_mgr_id" gorm:"type:uuid"`      // Manager who approved, foreign key, nullable

	// Timestamps
	DevApprovedAt *time.Time `json:"dev_approved_at"` // When developer approved, nullable
	MgrApprovedAt *time.Time `json:"mgr_approved_at"` // When manager approved, nullable

	// Relationships
	Bug       *Bug        `json:"bug,omitempty" gorm:"foreignKey:BugID;constraint:OnDelete:CASCADE"`
	Feedbacks []Feedback  `json:"feedbacks,omitempty" gorm:"foreignKey:ReleaseNoteID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID
func (rn *ReleaseNote) BeforeCreate(tx *gorm.DB) error {
	if rn.ID == uuid.Nil {
		rn.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for ReleaseNote model
func (ReleaseNote) TableName() string {
	return "release_notes"
}

