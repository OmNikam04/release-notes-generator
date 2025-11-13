package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/dto"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/models"
	"github.com/omnikam04/release-notes-generator/internal/repository"
	"github.com/omnikam04/release-notes-generator/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUser(id uuid.UUID) (*dto.UserResponse, error)
	UpdateUser(id uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id uuid.UUID) error
	Login(req *dto.LoginRequest) (*models.User, error)
	Logout(refreshToken string) error
	IssueRefreshToken(userID uuid.UUID) (string, error)
	RefreshTokens(refreshToken string) (*models.User, string, error)
}

type userService struct {
	userRepository    repository.UserRepository
	refreshRepository repository.RefreshTokenRepository
}

func NewUserService(userRepository repository.UserRepository, refreshRepository repository.RefreshTokenRepository) *userService {
	return &userService{userRepository: userRepository, refreshRepository: refreshRepository}
}

func (s *userService) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Check if user exists
	_, err := s.userRepository.FindByEmail(req.Email)
	if err == nil {
		// User found - duplicate email
		logger.Warn().Str("email", req.Email).Msg("User already exists")
		return nil, errors.New("user with this email already exists")
	}
	// If error is not "record not found", return the error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error().Err(err).Msg("Failed to check existing user")
		return nil, err
	}
	// User doesn't exist - proceed with creation

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to hash password")
		return nil, err
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepository.CreateUser(user); err != nil {
		logger.Error().Err(err).Msg("Failed to create user")
		return nil, err
	}

	logger.Info().Str("user_id", user.ID.String()).Msg("User created successfully")

	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) GetUser(id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("user_id", id.String()).Msg("User not found")
		return nil, errors.New("user not found")
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) UpdateUser(id uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("user_id", id.String()).Msg("User not found")
		return nil, errors.New("user not found")
	}

	// Update fields only if provided (not empty)
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// Hash new password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to hash password")
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	if err := s.userRepository.Update(user); err != nil {
		logger.Error().Err(err).Msg("Failed to update user")
		return nil, err
	}

	logger.Info().Str("user_id", user.ID.String()).Msg("User updated successfully")

	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) DeleteUser(id uuid.UUID) error {
	user, err := s.userRepository.FindByID(id)
	if err != nil {
		logger.Error().Err(err).Str("user_id", id.String()).Msg("User not found")
		return errors.New("user not found")
	}

	if err := s.userRepository.Delete(user.ID); err != nil {
		logger.Error().Err(err).Msg("Failed to delete user")
		return err
	}

	logger.Info().Str("user_id", user.ID.String()).Msg("User deleted successfully")
	return nil
}

func (s *userService) Login(req *dto.LoginRequest) (*models.User, error) {
	// Find user by email
	user, err := s.userRepository.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn().Str("email", req.Email).Msg("Login attempt with non-existent email")
			return nil, errors.New("invalid email or password")
		}
		logger.Error().Err(err).Msg("Failed to find user by email")
		return nil, errors.New("login failed")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		logger.Warn().Str("email", req.Email).Msg("Login attempt with incorrect password")
		return nil, errors.New("invalid email or password")
	}

	logger.Info().Str("user_id", user.ID.String()).Msg("User logged in successfully")
	return user, nil
}

// Logout revokes the provided refresh token, effectively logging out the user from that session
func (s *userService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token is required")
	}

	hash := utils.HashToken(refreshToken)
	rt, err := s.refreshRepository.FindByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Token doesn't exist - already logged out or invalid
			logger.Warn().Msg("Logout attempt with non-existent refresh token")
			return errors.New("invalid refresh token")
		}
		logger.Error().Err(err).Msg("Failed to find refresh token for logout")
		return err
	}

	// Revoke the refresh token
	if err := s.refreshRepository.Revoke(rt.ID); err != nil {
		logger.Error().Err(err).Str("token_id", rt.ID.String()).Msg("Failed to revoke refresh token")
		return errors.New("logout failed")
	}

	logger.Info().Str("user_id", rt.UserID.String()).Msg("User logged out successfully")
	return nil
}

// IssueRefreshToken generates and stores a new refresh token for the given user
func (s *userService) IssueRefreshToken(userID uuid.UUID) (string, error) {
	token, err := utils.GenerateSecureToken()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate refresh token")
		return "", err
	}
	hash := utils.HashToken(token)
	rt := &models.RefreshToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshRepository.Create(rt); err != nil {
		logger.Error().Err(err).Msg("Failed to persist refresh token")
		return "", err
	}
	return token, nil
}

// RefreshTokens validates the provided refresh token, rotates it, and returns the user + new refresh token
func (s *userService) RefreshTokens(refreshToken string) (*models.User, string, error) {
	if refreshToken == "" {
		return nil, "", errors.New("invalid refresh token")
	}
	hash := utils.HashToken(refreshToken)
	rt, err := s.refreshRepository.FindByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("invalid refresh token")
		}
		logger.Error().Err(err).Msg("Failed to find refresh token")
		return nil, "", err
	}
	if rt.RevokedAt != nil {
		logger.Warn().Msg("Attempt to use revoked refresh token")
		return nil, "", errors.New("refresh token revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		logger.Warn().Msg("Attempt to use expired refresh token")
		return nil, "", errors.New("refresh token expired")
	}

	user, err := s.userRepository.FindByID(rt.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("user not found")
		}
		return nil, "", err
	}

	// Rotate: revoke old and create new
	_ = s.refreshRepository.Revoke(rt.ID)

	newToken, err := utils.GenerateSecureToken()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate new refresh token")
		return nil, "", err
	}
	newHash := utils.HashToken(newToken)
	newRT := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshRepository.Create(newRT); err != nil {
		logger.Error().Err(err).Msg("Failed to persist new refresh token")
		return nil, "", err
	}

	return user, newToken, nil
}
