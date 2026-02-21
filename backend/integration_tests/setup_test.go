package integration_tests

import (
	"backend/internal/infra/postgres"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	BaseURL      string
	HTTPClient   *http.Client
	DbConnection *sql.DB
)

func TestMain(m *testing.M) {
	BaseURL = getEnv("TEST_BASE_URL", "http://localhost:8080")
	dbDSN := getEnv("TEST_DB_DSN", "postgres://devshare:devshare_password@localhost:5432/devshare?sslmode=disable")

	HTTPClient = &http.Client{Timeout: 10 * time.Second}

	if err := waitForApp(BaseURL + "/health"); err != nil {
		fmt.Fprintf(os.Stderr, "backend not ready: %v\n", err)
		os.Exit(1)
	}

	migrator, err := runMigrations(dbDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if migrator != nil {
		if err := migrator.Down(); err != nil && err != migrate.ErrNoChange {
			fmt.Fprintf(os.Stderr, "failed to rollback migrations: %v\n", err)
		}
		migrator.Close()
	}
	dbConfig := postgres.Config{
		Host: getEnv("TEST_DB_HOST", "localhost"),
		Port: 5432,
		User: getEnv("TEST_DB_USER",
			"postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnv("TEST_DB_NAME", "devshare"),
		SSLMode:  getEnv("TEST_DB_SSL_MODE", "disable"),
	}
	DbConnection, err = postgres.NewDB(dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	os.Exit(code)
}

func waitForApp(healthURL string) error {
	const maxAttempts = 30
	const delay = 2 * time.Second

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

func runMigrations(dbDSN string) (*migrate.Migrate, error) {
	migrator, err := migrate.New("file://../internal/infra/migrations", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("creating migrator: %w", err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return migrator, fmt.Errorf("running migrations: %w", err)
	}

	return migrator, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
