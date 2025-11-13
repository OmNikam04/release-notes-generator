package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/omnikam04/release-notes-generator/internal/config"
	"github.com/omnikam04/release-notes-generator/internal/dto"
	"github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/utils"
)

// Auth is a JWT authentication middleware
func Auth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warn().Msg("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing authorization token",
			})
		}

		// Extract token (remove "Bearer " prefix)
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Warn().Msg("Invalid Authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
		}

		tokenString := tokenParts[1]

		// Validate JWT token
		claims, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			logger.Warn().Err(err).Msg("Invalid JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
		}

		// Store user ID, email, and role in context for use in handlers
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRole", claims.Role)

		logger.Debug().
			Str("user_id", claims.UserID.String()).
			Str("email", claims.Email).
			Str("role", claims.Role).
			Msg("User authenticated successfully")

		// Call next handler
		return c.Next()
	}
}

// AuthMiddleware is an alias for Auth for consistency
func AuthMiddleware(jwtSecret string) fiber.Handler {
	cfg := &config.Config{JWTSecret: jwtSecret}
	return Auth(cfg)
}

// RoleMiddleware checks if the authenticated user has the required role
func RoleMiddleware(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user role from context (set by Auth middleware)
		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			logger.Warn().Msg("User role not found in context")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "User role not found",
			})
		}

		// Check if user has required role
		if userRole != requiredRole {
			logger.Warn().
				Str("user_role", userRole).
				Str("required_role", requiredRole).
				Msg("User does not have required role")
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "forbidden",
				Message: "You do not have permission to access this resource",
			})
		}

		// User has required role, proceed
		return c.Next()
	}
}
