package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/omnikam04/release-notes-generator/internal/api/handlers"
	"github.com/omnikam04/release-notes-generator/internal/api/routes"
	"github.com/omnikam04/release-notes-generator/internal/config"
	"github.com/omnikam04/release-notes-generator/internal/db"
	appLogger "github.com/omnikam04/release-notes-generator/internal/logger"
	"github.com/omnikam04/release-notes-generator/internal/repository"
	"github.com/omnikam04/release-notes-generator/internal/service"
)

func main() {
	// Initialize logger
	appLogger.Init("development")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Connect to database
	database, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Run database migrations
	// if err := db.RunMigrations(database); err != nil {
	// 	log.Fatalf("❌ Failed to run migrations: %v", err)
	// }

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	refreshRepo := repository.NewRefreshTokenRepository(database)

	// Initialize services
	userService := service.NewUserService(userRepo, refreshRepo)

	// Initialize handlers (pass config for JWT)
	userHandler := handlers.NewUserHandler(userService, cfg)

	// Create handlers struct for routing
	routeHandlers := &routes.Handlers{
		UserHandler: userHandler,
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Release notes generator API v1.0",
		DisableStartupMessage: false,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("❌ Error [%d]: %v | Path: %s | IP: %s", code, err, c.Path(), c.IP())
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))

	// Setup all routes (health, users, etc.)
	routes.SetupRoutes(app, routeHandlers, cfg)

	// Start server in a goroutine
	go func() {
		port := cfg.Port
		if port == "" {
			port = "8080"
		}
		// Bind to all interfaces
		listenAddr := fmt.Sprintf(":%s", port)
		log.Printf("🚀 Server starting on %s (accessible on all network interfaces)", listenAddr)
		if err := app.Listen(listenAddr); err != nil {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️  Shutting down server...")

	// Shutdown Fiber app
	if err := app.Shutdown(); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	// Close database connection
	if err := db.CloseDB(); err != nil {
		log.Printf("❌ Failed to close database: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
