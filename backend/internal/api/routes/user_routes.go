package routes

import (
"github.com/gofiber/fiber/v2"
"github.com/omnikam04/release-notes-generator/internal/api/middleware"
"github.com/omnikam04/release-notes-generator/internal/config"
)

// SetupUserRoutes sets up all user-related routes
func SetupUserRoutes(router fiber.Router, h *Handlers, cfg *config.Config) {
users := router.Group("/user")

// Public routes - no authentication required
users.Post("/login", h.UserHandler.Login)
users.Post("/refresh", h.UserHandler.RefreshTokens)
users.Post("/logout", h.UserHandler.Logout)

// Protected routes - require authentication
// Uses /me pattern - user can only access their own data
users.Get("/me", middleware.Auth(cfg), h.UserHandler.GetCurrentUser)
users.Delete("/me", middleware.Auth(cfg), h.UserHandler.DeleteCurrentUser)
}
