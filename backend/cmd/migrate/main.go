package main

import (
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dbPath := getEnv("DB_FILE_PATH", "./devshare.db")
	migrationsPath := getEnv("MIGRATIONS_PATH", "internal/infra/migrations/sqlite")

	m, err := migrate.New("file://"+migrationsPath, "sqlite://"+dbPath)
	if err != nil {
		slog.Error("migration init failed", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied successfully")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
