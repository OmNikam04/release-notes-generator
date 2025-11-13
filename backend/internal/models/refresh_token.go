package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken represents a persisted refresh token (stored as a hash)
type RefreshToken struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;index;not null"`
	TokenHash string    `json:"-" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	RevokedAt *time.Time `json:"revoked_at"`
}

func (t *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

