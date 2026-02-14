package main

import (
	"log"
	"os"
	"strconv"

	"backend/internal/infra/postgres"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	workspaceRepo := postgres.NewWorkspaceRepository(db)
	envRepo := postgres.NewEnvironmentRepository(db)

	// TODO: Pass repositories to handlers when handler layer is built
	_ = userRepo
	_ = workspaceRepo
	_ = envRepo

	app := fiber.New(fiber.Config{
		AppName: "Dev-Share Backend",
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

	// API routes placeholder
	api := app.Group("/api/v1")

	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Dev-Share API v1",
		})
	})

	// Get port from environment or default to 8080
	port := getEnv("PORT", "8080")

	log.Printf("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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
