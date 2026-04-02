package main

import (
	"log/slog"
	"os"

	"backend/internal/application"
	handlererrors "backend/internal/application/errors"
	"backend/internal/domain"
	"backend/internal/infra/filestorage"
	"backend/internal/infra/http/handlers"
	"backend/internal/infra/http/middleware"
	"backend/internal/infra/sqlite"
	"backend/internal/infra/terraform"
	"backend/internal/infra/tfparser"
	"backend/pkg/config"
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

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration loaded")

	// Database configuration
	dbConfig := sqlite.Config{
		FilePath: cfg.DBFilePath,
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
	jwtService, err := jwt.NewService(cfg.JWTSecret)
	if err != nil {
		slog.Error("failed to initialize JWT service", "error", err)
		os.Exit(1)
	}
	slog.Info("JWT service initialized")

	// File storage
	fileStorage := filestorage.NewLocalFileStorage(cfg.TemplateStoragePath)
	slog.Info("file storage initialized", "path", cfg.TemplateStoragePath)

	// Encryption
	encryptor, err := crypto.NewAESEncryptor(cfg.EncryptionKey)
	if err != nil {
		slog.Error("failed to initialize encryptor", "error", err)
		os.Exit(1)
	}
	slog.Info("encryption service initialized")

	// TF Parser
	tfParser := tfparser.NewHCLParser()
	// Execution storage for terraform working directories
	executionStorage := filestorage.NewLocalExecutionStorage(cfg.EnvExecutionPath, cfg.TemplateStoragePath)
	slog.Info("execution storage initialized", "path", cfg.EnvExecutionPath)

	// Terraform executor
	tfExecutor := terraform.NewExecutor(cfg.EnvExecutionPath, cfg.TFPluginCacheDir)
	slog.Info("terraform executor initialized")

	// Infrastructure factories
	uowFactory := sqlite.NewUnitOfWorkFactory(db)
	repoFactory := sqlite.NewRepositoryFactory()

	// Application-layer service factory
	serviceFactory := application.NewServiceFactory(uowFactory, repoFactory, validator, fileStorage, encryptor, tfParser, executionStorage, tfExecutor)

	// Initialize handlers — method values satisfy the handler's func() (Service, UnitOfWork) field
	userHandler := handlers.NewUserHandler(serviceFactory.NewUserService)
	workspaceHandler := handlers.NewWorkspaceHandler(serviceFactory.NewWorkspaceService)
	templateHandler := handlers.NewTemplateHandler(serviceFactory.NewTemplateService)
	templateVariableHandler := handlers.NewTemplateVariableHandler(serviceFactory.NewTemplateVariableService)
	envVarValueHandler := handlers.NewEnvironmentVariableValueHandler(serviceFactory.NewEnvironmentVariableValueService)
	environmentHandler := handlers.NewEnvironmentHandler(serviceFactory.NewEnvironmentService)
	adminHandler := handlers.NewAdminHandler(serviceFactory.NewAdminService)

	app := fiber.New(fiber.Config{
		AppName:      "Dev-Share Backend",
		ErrorHandler: handlererrors.ErrorHandler(),
		BodyLimit:    cfg.BodyLimitBytes,
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
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
	environmentHandler.RegisterRoutes(protected)

	// Get port from environment or default to 8080
	port := getEnv("PORT", "8080")

	slog.Info("starting server", "port", port)
	if err := app.Listen(":" + port); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
