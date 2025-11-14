package main

import (
	"context"
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
	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/external/gemini"
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

	// Run database migrations (only if RUN_MIGRATIONS=true)
	runMigrations := os.Getenv("RUN_MIGRATIONS")
	if runMigrations == "true" {
		appLogger.Info().Msg("🔄 Running database migrations...")
		if err := db.RunMigrations(database); err != nil {
			log.Fatalf("❌ Failed to run migrations: %v", err)
		}
		appLogger.Info().Msg("✅ Database migrations completed successfully")
	} else {
		appLogger.Info().Msg("⏭️  Skipping migrations (RUN_MIGRATIONS not set to 'true')")
	}

	// Initialize Bugsby client
	bugsbyClient, err := bugsby.NewClient(&bugsby.Config{
		BaseURL:   cfg.BugsbyAPIURL,
		TokenFile: cfg.BugsbyTokenFile,
	})
	if err != nil {
		log.Fatalf("❌ Failed to initialize Bugsby client: %v", err)
	}
	appLogger.Info().Msg("✅ Bugsby client initialized successfully")

	// Initialize AI service (Gemini)
	var aiService service.AIService
	appLogger.Info().
		Str("gcp_project_id", cfg.GCPProjectID).
		Str("gcp_location", cfg.GCPLocation).
		Str("gemini_model", cfg.GeminiModel).
		Msg("🔍 Checking AI service configuration")

	if cfg.GCPProjectID != "" && cfg.GCPLocation != "" {
		appLogger.Info().Msg("🚀 Initializing AI service (Gemini)...")
		ctx := context.Background()
		aiService, err = service.NewAIService(ctx, &gemini.Config{
			ProjectID: cfg.GCPProjectID,
			Location:  cfg.GCPLocation,
			Model:     cfg.GeminiModel,
		})
		if err != nil {
			appLogger.Warn().Err(err).Msg("⚠️  Failed to initialize AI service, will use placeholder generation")
			aiService = nil
		} else {
			appLogger.Info().
				Str("model", cfg.GeminiModel).
				Msg("✅ AI service (Gemini) initialized successfully")
		}
	} else {
		appLogger.Warn().
			Str("gcp_project_id", cfg.GCPProjectID).
			Str("gcp_location", cfg.GCPLocation).
			Msg("⚠️  AI service not configured (missing GCP credentials), will use placeholder generation")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	refreshRepo := repository.NewRefreshTokenRepository(database)
	bugRepo := repository.NewBugRepository(database)
	releaseNoteRepo := repository.NewReleaseNoteRepository(database)
	feedbackRepo := repository.NewFeedbackRepository(database)
	patternRepo := repository.NewPatternRepository(database)
	feedbackPatternRepo := repository.NewFeedbackPatternRepository(database)

	// Initialize services
	userService := service.NewUserService(userRepo, refreshRepo)
	bugsbySyncService := service.NewBugsbySyncService(bugsbyClient, bugRepo, userRepo)

	// Initialize feedback and pattern services
	var feedbackService service.FeedbackService
	var patternService service.PatternService

	if aiService != nil && cfg.GCPProjectID != "" && cfg.GCPLocation != "" {
		// Create a separate Gemini client for pattern service
		ctx := context.Background()
		geminiClient, err := gemini.NewClient(ctx, &gemini.Config{
			ProjectID: cfg.GCPProjectID,
			Location:  cfg.GCPLocation,
			Model:     cfg.GeminiModel,
		})
		if err != nil {
			appLogger.Warn().Err(err).Msg("⚠️  Failed to create Gemini client for pattern service")
		} else {
			// Pattern service needs Gemini client for pattern extraction
			patternService = service.NewPatternService(patternRepo, feedbackRepo, feedbackPatternRepo, geminiClient)
			feedbackService = service.NewFeedbackService(feedbackRepo, bugRepo, patternService)
			appLogger.Info().Msg("✅ Feedback and pattern services initialized")
		}
	} else {
		// If no AI service, create nil services (won't capture feedback)
		appLogger.Warn().Msg("⚠️  Feedback and pattern services disabled (no AI service)")
	}

	releaseNoteService := service.NewReleaseNoteService(releaseNoteRepo, bugRepo, bugsbyClient, aiService, feedbackService, database)

	// Initialize handlers (pass config for JWT)
	userHandler := handlers.NewUserHandler(userService, cfg)
	bugHandler := handlers.NewBugHandler(bugsbySyncService, bugRepo, bugsbyClient, releaseNoteService)
	releaseNoteHandler := handlers.NewReleaseNoteHandler(releaseNoteService)

	// Create handlers struct for routing
	routeHandlers := &routes.Handlers{
		UserHandler:        userHandler,
		BugHandler:         bugHandler,
		ReleaseNoteHandler: releaseNoteHandler,
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
