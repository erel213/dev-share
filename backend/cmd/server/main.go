package main

import (
	"log/slog"
	"os"

	"backend/internal/application"
	handlererrors "backend/internal/application/errors"
	"backend/internal/infra/http/handlers"
	"backend/internal/infra/http/middleware"
	"backend/internal/infra/sqlite"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Initialize structured logging
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(slogger)

	slog.Info("starting dev-share backend")

	// Database configuration from environment variables
	dbConfig := sqlite.Config{
		FilePath: getEnv("DB_FILE_PATH", "./devshare.db"),
	}

	// Initialize database connection
	db, err := sqlite.NewDB(dbConfig)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("successfully connected to database")

	// Run migrations
	// TODO: Consider moving the logic for running migrations to onboarding package
	migrationsPath := getEnv("MIGRATIONS_PATH", "internal/infra/migrations/sqlite")
	m, err := migrate.New(
		"file://"+migrationsPath,
		"sqlite://"+dbConfig.FilePath,
	)
	if err != nil {
		slog.Error("migration init failed", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// Initialize validation service
	validator := validation.New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		slog.Error("failed to register custom validations", "error", err)
		os.Exit(1)
	}
	slog.Info("validation service initialized")

	// Initialize JWT service
	jwtService, err := jwt.NewService()
	if err != nil {
		slog.Error("failed to initialize JWT service", "error", err)
		os.Exit(1)
	}
	slog.Info("JWT service initialized")

	// Infrastructure factories
	uowFactory := sqlite.NewUnitOfWorkFactory(db)
	repoFactory := sqlite.NewRepositoryFactory()

	// Application-layer service factory
	serviceFactory := application.NewServiceFactory(uowFactory, repoFactory, validator)

	// Initialize handlers — method values satisfy the handler's func() (Service, UnitOfWork) field
	userHandler := handlers.NewUserHandler(serviceFactory.NewUserService)
	workspaceHandler := handlers.NewWorkspaceHandler(serviceFactory.NewWorkspaceService)
	templateHandler := handlers.NewTemplateHandler(serviceFactory.NewTemplateService)
	adminHandler := handlers.NewAdminHandler(serviceFactory.NewAdminService)

	app := fiber.New(fiber.Config{
		AppName:      "Dev-Share Backend",
		ErrorHandler: handlererrors.ErrorHandler(),
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "healthy",
			"service":  "dev-share-backend",
			"database": "connected",
		})
	})

	// Admin initialization endpoint (unprotected, first-time only)
	app.Post("/admin/init", adminHandler.InitializeSystem)

	// API routes
	api := app.Group("/api/v1")

	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Dev-Share API v1",
		})
	})

	// Public: user registration does not require authentication
	userHandler.RegisterRoutes(api)

	protected := api.Group("", middleware.RequireAuth(jwtService, jwt.DefaultCookieConfig()))
	workspaceHandler.RegisterRoutes(protected)
	templateHandler.RegisterRoutes(protected)

	// Get port from environment or default to 8080
	port := getEnv("PORT", "8080")

	slog.Info("starting server", "port", port)
	if err := app.Listen(":" + port); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
