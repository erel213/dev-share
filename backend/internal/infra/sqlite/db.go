package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Config struct {
	FilePath string
}

func NewDB(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", cfg.FilePath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite serializes writes; a single connection avoids lock contention.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	// Verify foreign keys are enabled.
	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		return nil, fmt.Errorf("failed to check foreign_keys pragma: %w", err)
	}
	if fkEnabled != 1 {
		return nil, fmt.Errorf("foreign key enforcement is not enabled")
	}

	return db, nil
}
