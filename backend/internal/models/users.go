package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"unique;not null;index"`
	Password string `json:"-" gorm:"not null"` // "-" hides password from JSON responses
}

// BeforeCreate hook to generate UUID before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName overrides the default table name
func (User) TableName() string {
	return "users"
}
