package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByHash(hash string) (*models.RefreshToken, error)
	Revoke(id uuid.UUID) error
	RevokeAllForUser(userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByHash(hash string) (*models.RefreshToken, error) {
	var t models.RefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&t).Error
	return &t, err
}

func (r *refreshTokenRepository) Revoke(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).Where("id = ?", id).Update("revoked_at", now).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}
