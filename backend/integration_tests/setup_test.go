package integration_tests

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"backend/internal/application"
	handlererrors "backend/internal/application/errors"
	"backend/internal/infra/http/handlers"
	"backend/internal/infra/http/middleware"
	"backend/internal/infra/sqlite"
	"backend/pkg/jwt"
	"backend/pkg/validation"

	"github.com/gofiber/fiber/v2"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	BaseURL      string
	HTTPClient   *http.Client
	DbConnection *sql.DB
	jwtSvc       *jwt.Service
)

func TestMain(m *testing.M) {
	// Ensure JWT_SECRET is set for the test process.
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "your_jwt_secretyour_jwt_secretyour_jwt_secretyour_jwt_secret")
	}

	HTTPClient = &http.Client{Timeout: 10 * time.Second}

	var err error
	jwtSvc, err = jwt.NewService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create JWT service: %v\n", err)
		os.Exit(1)
	}

	// Create a temporary directory for the SQLite database.
	tmpDir, err := os.MkdirTemp("", "devshare-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	dbFilePath := filepath.Join(tmpDir, "devshare.db")

	// Open the database (shared with the in-process backend).
	DbConnection, err = sqlite.NewDB(sqlite.Config{FilePath: dbFilePath})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer DbConnection.Close()

	// Run migrations.
	migrator, err := migrate.New(
		"file://../internal/infra/migrations/sqlite",
		"sqlite://"+dbFilePath,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrator: %v\n", err)
		os.Exit(1)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Set up application services.
	validator := validation.New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register validations: %v\n", err)
		os.Exit(1)
	}

	uowFactory := sqlite.NewUnitOfWorkFactory(DbConnection)
	repoFactory := sqlite.NewRepositoryFactory()
	serviceFactory := application.NewServiceFactory(uowFactory, repoFactory, validator)

	// Build the Fiber app (mirrors cmd/server/main.go).
	app := fiber.New(fiber.Config{
		AppName:      "Dev-Share Backend Test",
		ErrorHandler: handlererrors.ErrorHandler(),
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	adminHandler := handlers.NewAdminHandler(serviceFactory.NewAdminService)
	app.Post("/admin/init", adminHandler.InitializeSystem)

	api := app.Group("/api/v1")

	userHandler := handlers.NewUserHandler(serviceFactory.NewUserService)
	userHandler.RegisterRoutes(api)

	protected := api.Group("", middleware.RequireAuth(jwtSvc, jwt.DefaultCookieConfig()))
	workspaceHandler := handlers.NewWorkspaceHandler(serviceFactory.NewWorkspaceService)
	workspaceHandler.RegisterRoutes(protected)

	templateHandler := handlers.NewTemplateHandler(serviceFactory.NewTemplateService)
	templateHandler.RegisterRoutes(protected)

	// Listen on a random available port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}
	BaseURL = fmt.Sprintf("http://%s", ln.Addr().String())

	go func() {
		if err := app.Listener(ln); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	// Wait for the server to accept connections.
	if err := waitForApp(BaseURL + "/health"); err != nil {
		fmt.Fprintf(os.Stderr, "backend not ready: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	app.Shutdown()
	if migrator != nil {
		migrator.Down()
		migrator.Close()
	}

	os.Exit(code)
}

func waitForApp(healthURL string) error {
	const maxAttempts = 30
	const delay = 200 * time.Millisecond

	for i := range maxAttempts {
		resp, err := HTTPClient.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		fmt.Printf("waiting for backend (attempt %d/%d)...\n", i+1, maxAttempts)
		time.Sleep(delay)
	}

	return fmt.Errorf("backend did not become ready after %d attempts", maxAttempts)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
