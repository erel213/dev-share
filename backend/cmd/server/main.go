package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"backend/internal/application"
	handlererrors "backend/internal/application/errors"
	"backend/internal/infra/http/handlers"
	"backend/internal/infra/http/middleware"
	"backend/internal/infra/postgres"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Initialize structured logging
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(slogger)

	slog.Info("starting dev-share backend")

	// Database configuration from environment variables
	dbConfig := postgres.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "devshare"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	// Initialize database connection
	db, err := postgres.NewDB(dbConfig)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("successfully connected to database")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	workspaceRepo := postgres.NewWorkspaceRepository(db)
	envRepo := postgres.NewEnvironmentRepository(db)

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

	// Initialize services
	userService := application.NewUserService(userRepo, validator)
	workspaceService := application.NewWorkspaceService(workspaceRepo, validator)

	cookieCfg := jwt.DefaultCookieConfig()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, jwtService, cookieCfg)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService)

	_ = envRepo

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

	// API routes
	api := app.Group("/api/v1")

	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Dev-Share API v1",
		})
	})

	// Public: user registration does not require authentication
	userHandler.RegisterRoutes(api)

	protected := api.Group("", middleware.RequireAuth(jwtService, cookieCfg))
	workspaceHandler.RegisterRoutes(protected)

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

// getEnvInt retrieves an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid integer value for %s: %v, using default %d", key, err, defaultValue)
		return defaultValue
	}
	return value
}
