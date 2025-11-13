package routes

import (
	"github.com/gofiber/fiber/v2"
)

// SetupHealthRoutes sets up health check and root routes
// These routes don't have /api prefix
func SetupHealthRoutes(app *fiber.App) {
	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		// Debug logging
		println("🔍 Health check request from:", c.IP(), "| Host:", c.Hostname(), "| Protocol:", c.Protocol())

		// Try to send response
		err := c.JSON(fiber.Map{
			"status":  "ok",
			"service": "release-notes-generator",
			"version": "1.0.0",
		})

		if err != nil {
			println("❌ Error sending JSON response:", err.Error())
		}

		return err
	})

	// Root endpoint - API information
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Bike Health Tracker API",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"health": "/health",
				"api":    "/api/v1",
			},
		})
	})
}
