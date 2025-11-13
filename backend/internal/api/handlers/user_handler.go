package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/omnikam04/release-notes-generator/internal/config"
	"github.com/omnikam04/release-notes-generator/internal/dto"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/service"
	"github.com/omnikam04/release-notes-generator/internal/utils"
)

type UserHandler struct {
	userService service.UserService
	config      *config.Config
}

func NewUserHandler(userService service.UserService, cfg *config.Config) *UserHandler {
	return &UserHandler{
		userService: userService,
		config:      cfg,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := ValidateStruct(c, &req); err != nil {
		return err
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
	}

	logger.Info().Interface("user", user).Msg("User created via API")

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse{
		Success: true,
		Data:    user,
		Message: "User created successfully",
	})
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Tags users
// @Produce json
// @Success 200 {object} dto.SuccessResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	// Extract authenticated user ID from JWT context (set by Auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		logger.Error().Msg("Failed to extract userID from context")
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    user,
	})
}

// UpdateCurrentUser godoc
// @Summary Update current user profile
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UpdateUserRequest true "User data to update"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateCurrentUser(c *fiber.Ctx) error {
	// Extract authenticated user ID from JWT context (set by Auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		logger.Error().Msg("Failed to extract userID from context")
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := ValidateStruct(c, &req); err != nil {
		return err
	}

	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    user,
		Message: "User updated successfully",
	})
}

// DeleteCurrentUser godoc
// @Summary Delete current user account
// @Tags users
// @Produce json
// @Success 200 {object} dto.SuccessResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/me [delete]
func (h *UserHandler) DeleteCurrentUser(c *fiber.Ctx) error {
	// Extract authenticated user ID from JWT context (set by Auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		logger.Error().Msg("Failed to extract userID from context")
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user context",
		})
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// Login godoc
// @Summary User login
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /users/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := ValidateStruct(c, &req); err != nil {
		return err
	}

	user, err := h.userService.Login(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "login_failed",
			Message: err.Error(),
		})
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, h.config.JWTSecret)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate JWT token")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate authentication token",
		})
	}

	logger.Info().Str("user_id", user.ID.String()).Msg("User logged in successfully")
	// Issue refresh token
	refreshToken, err := h.userService.IssueRefreshToken(user.ID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to generate refresh token")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate refresh token",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data: dto.LoginResponse{
			Token:        token,
			RefreshToken: refreshToken,
			User: dto.UserResponse{
				ID:        user.ID,
				Name:      user.Name,
				Email:     user.Email,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			},
		},
		Message: "Login successful",
	})
}

// RefreshTokens godoc
// @Summary Refresh access token
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /user/refresh [post]
func (h *UserHandler) RefreshTokens(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body for refresh")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := ValidateStruct(c, &req); err != nil {
		return err
	}

	user, newRefreshToken, err := h.userService.RefreshTokens(req.RefreshToken)
	if err != nil {
		logger.Warn().Err(err).Msg("Refresh token invalid or expired")
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "refresh_failed",
			Message: err.Error(),
		})
	}

	newAccessToken, err := utils.GenerateToken(user.ID, user.Email, h.config.JWTSecret)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate new access token")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate authentication token",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data: dto.RefreshTokenResponse{
			Token:        newAccessToken,
			RefreshToken: newRefreshToken,
		},
	})
}

// Logout godoc
// @Summary User logout
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.RefreshTokenRequest true "Refresh token to revoke"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /user/logout [post]
func (h *UserHandler) Logout(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request body for logout")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := ValidateStruct(c, &req); err != nil {
		return err
	}

	if err := h.userService.Logout(req.RefreshToken); err != nil {
		logger.Warn().Err(err).Msg("Logout failed")
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "logout_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}
