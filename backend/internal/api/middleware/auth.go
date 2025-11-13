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
