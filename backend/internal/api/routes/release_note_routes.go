package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/omnikam04/release-notes-generator/internal/api/middleware"
	"github.com/omnikam04/release-notes-generator/internal/config"
)

// SetupReleaseNoteRoutes sets up all release note-related routes
func SetupReleaseNoteRoutes(router fiber.Router, h *Handlers, cfg *config.Config) {
	// Release notes group - all routes require authentication
	releaseNotes := router.Group("/release-notes")
	releaseNotes.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Endpoint 1: Get bugs WITH release notes (Kanban view)
	// GET /api/v1/release-notes?assigned_to_me=true&status=ai_generated
	releaseNotes.Get("/", h.ReleaseNoteHandler.GetReleaseNotes)

	// Endpoint 2: Get bugs without release notes (pending)
	// GET /api/v1/release-notes/pending?assigned_to_me=true&release=wifi-ooty
	releaseNotes.Get("/pending", h.ReleaseNoteHandler.GetPendingBugs)

	// Endpoint 3: Get bug context with commit information
	// GET /api/v1/release-notes/bug/:bug_id/context
	releaseNotes.Get("/bug/:bug_id/context", h.ReleaseNoteHandler.GetBugContext)

	// Endpoint 4: Generate release note
	// POST /api/v1/release-notes/generate
	releaseNotes.Post("/generate", h.ReleaseNoteHandler.GenerateReleaseNote)

	// Endpoint 5: Get release note by bug ID
	// GET /api/v1/release-notes/bug/:bug_id
	releaseNotes.Get("/bug/:bug_id", h.ReleaseNoteHandler.GetReleaseNoteByBugID)

	// Endpoint 6: Update release note
	// PUT /api/v1/release-notes/:id
	releaseNotes.Put("/:id", h.ReleaseNoteHandler.UpdateReleaseNote)

	// Endpoint 7: Bulk generate release notes
	// POST /api/v1/release-notes/bulk-generate
	releaseNotes.Post("/bulk-generate", h.ReleaseNoteHandler.BulkGenerateReleaseNotes)

	// Manager-only endpoints
	managerRoutes := releaseNotes.Group("")
	managerRoutes.Use(middleware.RoleMiddleware("manager"))

	// Endpoint 8: Approve/reject release note (manager only)
	// POST /api/v1/release-notes/:id/approve
	managerRoutes.Post("/:id/approve", h.ReleaseNoteHandler.ApproveReleaseNote)
}
