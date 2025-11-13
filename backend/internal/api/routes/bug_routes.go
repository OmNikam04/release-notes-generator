package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/omnikam04/release-notes-generator/internal/api/middleware"
	"github.com/omnikam04/release-notes-generator/internal/config"
)

// SetupBugRoutes sets up all bug-related routes
func SetupBugRoutes(router fiber.Router, h *Handlers, cfg *config.Config) {
	// Bugsby sync endpoints (manager only)
	bugsby := router.Group("/bugsby")
	bugsby.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	bugsby.Use(middleware.RoleMiddleware("manager")) // Only managers can sync
	bugsby.Post("/sync", h.BugHandler.SyncRelease)
	bugsby.Post("/sync/:bugsby_id", h.BugHandler.SyncBugByID)
	bugsby.Get("/status", h.BugHandler.GetSyncStatus)

	// Bug management endpoints
	bugs := router.Group("/bugs")
	bugs.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	
	// All authenticated users can view bugs
	bugs.Get("/", h.BugHandler.ListBugs)
	bugs.Get("/:id", h.BugHandler.GetBug)
	
	// Only managers can update/delete bugs
	bugs.Patch("/:id", middleware.RoleMiddleware("manager"), h.BugHandler.UpdateBug)
	bugs.Delete("/:id", middleware.RoleMiddleware("manager"), h.BugHandler.DeleteBug)
}

