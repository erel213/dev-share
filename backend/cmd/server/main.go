package main

import (
	"encoding/hex"
	"log/slog"
	"os"

	"backend/internal/application"
	handlererrors "backend/internal/application/errors"
	"backend/internal/infra/filestorage"
	"backend/internal/infra/http/handlers"
	"backend/internal/infra/http/middleware"
	"backend/internal/infra/sqlite"
	"backend/internal/infra/tfparser"
	"backend/pkg/crypto"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// File storage
	templateStoragePath := getEnv("TEMPLATE_STORAGE_PATH", "./template_storage")
	fileStorage := filestorage.NewLocalFileStorage(templateStoragePath)
	slog.Info("file storage initialized", "path", templateStoragePath)

	// Encryption
	encryptionKeyHex := getEnv("ENCRYPTION_KEY", "")
	if encryptionKeyHex == "" {
		slog.Error("ENCRYPTION_KEY environment variable is required")
		os.Exit(1)
	}
	encryptionKey, err := hex.DecodeString(encryptionKeyHex)
	if err != nil {
		slog.Error("ENCRYPTION_KEY must be a valid hex string", "error", err)
		os.Exit(1)
	}
	encryptor, err := crypto.NewAESEncryptor(encryptionKey)
	if err != nil {
		slog.Error("failed to initialize encryptor", "error", err)
		os.Exit(1)
	}
	slog.Info("encryption service initialized")

	// TF Parser
	tfParser := tfparser.NewHCLParser()

	// Infrastructure factories
	uowFactory := sqlite.NewUnitOfWorkFactory(db)
	repoFactory := sqlite.NewRepositoryFactory()

	// Application-layer service factory
	serviceFactory := application.NewServiceFactory(uowFactory, repoFactory, validator, fileStorage, encryptor, tfParser)

	// Initialize handlers — method values satisfy the handler's func() (Service, UnitOfWork) field
	userHandler := handlers.NewUserHandler(serviceFactory.NewUserService)
	workspaceHandler := handlers.NewWorkspaceHandler(serviceFactory.NewWorkspaceService)
	templateHandler := handlers.NewTemplateHandler(serviceFactory.NewTemplateService)
	templateVariableHandler := handlers.NewTemplateVariableHandler(serviceFactory.NewTemplateVariableService)
	envVarValueHandler := handlers.NewEnvironmentVariableValueHandler(serviceFactory.NewEnvironmentVariableValueService)
	adminHandler := handlers.NewAdminHandler(serviceFactory.NewAdminService)

	app := fiber.New(fiber.Config{
		AppName:      "Dev-Share Backend",
		ErrorHandler: handlererrors.ErrorHandler(),
		BodyLimit:    10 * 1024 * 1024, // 10MB
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:3000",
		AllowCredentials: true,
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "healthy",
			"service":  "dev-share-backend",
			"database": "connected",
		})
	})

	// Admin endpoints (unprotected, first-time only)
	app.Get("/admin/status", adminHandler.GetSystemStatus)
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
	userHandler.RegisterProtectedRoutes(protected)
	workspaceHandler.RegisterRoutes(protected)
	templateHandler.RegisterRoutes(protected)
	templateVariableHandler.RegisterRoutes(protected)
	envVarValueHandler.RegisterRoutes(protected)

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
