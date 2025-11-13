package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Feedback represents manager feedback on AI-generated release notes for learning
type Feedback struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	ReleaseNoteID uuid.UUID `json:"release_note_id" gorm:"type:uuid;not null;index"` // Foreign key to release_notes
	BugID         uuid.UUID `json:"bug_id" gorm:"type:uuid;not null;index"`          // Foreign key to bugs (denormalized for querying)
	ManagerID     uuid.UUID `json:"manager_id" gorm:"type:uuid;not null;index"`      // Foreign key to users (who gave feedback)

	// Feedback Content (What manager provided)
	OriginalContent  string  `json:"original_content" gorm:"type:text;not null"`  // AI-generated content
	CorrectedContent string  `json:"corrected_content" gorm:"type:text;not null"` // Manager's corrected version
	FeedbackText     *string `json:"feedback_text" gorm:"type:text"`              // Natural language feedback (nullable)

	// AI-Extracted Patterns (Result of pattern extraction prompt)
	ExtractedPatterns datatypes.JSON `json:"extracted_patterns" gorm:"type:jsonb;not null;default:'{}'"`     // Array of pattern objects
	OverallConfidence float64        `json:"overall_confidence" gorm:"type:decimal(3,2);not null;default:0"` // Overall confidence from AI

	// Bug Context (Snapshot at time of feedback - for similarity matching)
	BugContext datatypes.JSON `json:"bug_context" gorm:"type:jsonb;not null;default:'{}'"`
	// Example: {
	//   "bug_type": "security",
	//   "severity": "high",
	//   "has_cve": true,
	//   "cve_number": "CVE-2025-32990",
	//   "component": "gnutls",
	//   "title_keywords": ["vulnerability", "security", "gnutls"]
	// }

	// Action Taken
	Action string `json:"action" gorm:"type:varchar(50);not null"` // "approved_with_correction", "sent_back_to_dev"

	// Learning Metrics
	TimesUsedAsExample int      `json:"times_used_as_example" gorm:"default:0"`           // How many times used in few-shot
	EffectivenessScore *float64 `json:"effectiveness_score" gorm:"type:decimal(3,2)"`     // 0.0-1.0, nullable (calculated later)

	// Pattern Processing Status
	PatternsExtracted bool    `json:"patterns_extracted" gorm:"default:false"` // Has AI extracted patterns yet?
	ExtractionError   *string `json:"extraction_error" gorm:"type:text"`       // Error if extraction failed

	// Relationships
	ReleaseNote      *ReleaseNote       `json:"release_note,omitempty" gorm:"foreignKey:ReleaseNoteID;constraint:OnDelete:CASCADE"`
	Bug              *Bug               `json:"bug,omitempty" gorm:"foreignKey:BugID;constraint:OnDelete:CASCADE"`
	Manager          *User              `json:"manager,omitempty" gorm:"foreignKey:ManagerID;constraint:OnDelete:SET NULL"`
	FeedbackPatterns []FeedbackPattern  `json:"feedback_patterns,omitempty" gorm:"foreignKey:FeedbackID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID
func (f *Feedback) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for Feedback model
func (Feedback) TableName() string {
	return "feedbacks"
}

