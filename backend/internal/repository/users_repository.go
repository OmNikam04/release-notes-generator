package repository

import (
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"gorm.io/gorm"
)

/*
This defines what operations are supported for User and does not care how they are implemented.
Any type that implements all these methods is a UserRepository.
*/
type UserRepository interface {
	CreateUser(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
}

/*
UserRepository is a concrete implementation of UserRepositoryInterface.
  - An interface says what should be done.
  - A concrete implementation says how it’s done.

Actual working code that fulfills the contract defined by an interface.
It’s the actual implementation of UserRepository interface.
*/
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}
