package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bug represents a bug from Bugsby system
type Bug struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Bugsby Integration
	BugsbyID  string `json:"bugsby_id" gorm:"type:varchar(50);uniqueIndex;not null"` // Bug ID from Bugsby (e.g., "1257310")
	BugsbyURL string `json:"bugsby_url" gorm:"type:varchar(500)"`                     // Full URL to bug in Bugsby

	// Bug Details
	Title       string  `json:"title" gorm:"type:text;not null"`         // Bug title/summary
	Description *string `json:"description" gorm:"type:text"`            // Full bug description (nullable)
	Severity    string  `json:"severity" gorm:"type:varchar(20);index"`  // "critical", "high", "medium", "low"
	Priority    string  `json:"priority" gorm:"type:varchar(10)"`        // "P0", "P1", "P2", "P3"
	BugType     string  `json:"bug_type" gorm:"type:varchar(50);index"`  // "security", "feature", "bugfix", "enhancement"
	CVENumber   *string `json:"cve_number" gorm:"type:varchar(50)"`      // CVE number if security bug (nullable)

	// Assignment
	AssignedTo *uuid.UUID `json:"assigned_to" gorm:"type:uuid;index"` // Developer user ID (nullable, foreign key)
	ManagerID  *uuid.UUID `json:"manager_id" gorm:"type:uuid;index"`  // Manager user ID (nullable, foreign key)

	// Release Info
	Release   string `json:"release" gorm:"type:varchar(100);not null;index"` // Release name (e.g., "wifi-ooty")
	Component string `json:"component" gorm:"type:varchar(100);index"`        // Component name (e.g., "gnutls", "CAS-ALMA9")

	// Status Tracking
	Status string `json:"status" gorm:"type:varchar(50);not null;index;default:'pending'"` // "pending", "ai_generated", "dev_approved", "mgr_approved", "rejected"

	// Bugsby Sync
	LastSyncedAt *time.Time `json:"last_synced_at"` // Last time synced from Bugsby (nullable)
	SyncStatus   string     `json:"sync_status" gorm:"type:varchar(20);default:'pending'"` // "synced", "pending", "failed"

	// Relationships
	ReleaseNote *ReleaseNote `json:"release_note,omitempty" gorm:"foreignKey:BugID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID
func (b *Bug) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for Bug model
func (Bug) TableName() string {
	return "bugs"
}

