package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditLog tracks all changes for accountability and analytics
type AuditLog struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `json:"created_at"` // No UpdatedAt/DeletedAt - audit logs are immutable

	// What happened
	EntityType string    `json:"entity_type" gorm:"type:varchar(50);not null;index"` // "bug", "release_note", "feedback", "pattern"
	EntityID   uuid.UUID `json:"entity_id" gorm:"type:uuid;not null;index"`          // ID of the entity
	Action     string    `json:"action" gorm:"type:varchar(50);not null;index"`      // "created", "updated", "approved", "rejected", "synced", "regenerated"

	// Who did it
	UserID    *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`      // NULL for system actions, foreign key
	UserEmail string     `json:"user_email" gorm:"type:varchar(255)"` // Denormalized for easy display
	UserRole  string     `json:"user_role" gorm:"type:varchar(50)"`   // "developer", "manager", "system"

	// Details
	Changes  datatypes.JSON `json:"changes" gorm:"type:jsonb"`  // What changed (before/after values)
	Metadata datatypes.JSON `json:"metadata" gorm:"type:jsonb"` // Additional context (e.g., AI confidence, time spent)

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// BeforeCreate hook to generate UUID
func (al *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

