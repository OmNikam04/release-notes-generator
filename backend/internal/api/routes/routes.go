package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/omnikam04/release-notes-generator/internal/api/handlers"
	"github.com/omnikam04/release-notes-generator/internal/config"
)

// Handlers struct holds all handler instances
type Handlers struct {
	UserHandler *handlers.UserHandler
	BugHandler  *handlers.BugHandler
}

// SetupRoutes registers all application routes
func SetupRoutes(app *fiber.App, handlers *Handlers, cfg *config.Config) {
	// Health check routes (no /api prefix)
	SetupHealthRoutes(app)

	// API v1 group
	api := app.Group("/api/v1")

	// Register resource-specific routes
	SetupUserRoutes(api, handlers, cfg)
	SetupBugRoutes(api, handlers, cfg)
}
