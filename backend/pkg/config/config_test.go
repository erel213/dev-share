package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnvOrFile_FileTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret")
	if err := os.WriteFile(path, []byte("from-file"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_SECRET", "from-env")
	t.Setenv("TEST_SECRET_FILE", path)

	got, err := getEnvOrFile("TEST_SECRET", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from-file" {
		t.Errorf("want %q, got %q", "from-file", got)
	}
}

func TestGetEnvOrFile_FileMissingReturnsError(t *testing.T) {
	t.Setenv("TEST_SECRET_FILE", "/nonexistent/path/secret")

	_, err := getEnvOrFile("TEST_SECRET", "default")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestGetEnvOrFile_FallsBackToEnv(t *testing.T) {
	t.Setenv("TEST_SECRET", "from-env")

	got, err := getEnvOrFile("TEST_SECRET", "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from-env" {
		t.Errorf("want %q, got %q", "from-env", got)
	}
}

func TestGetEnvOrFile_FallsBackToDefault(t *testing.T) {
	got, err := getEnvOrFile("TEST_SECRET_UNSET_XYZ", "fallback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "fallback" {
		t.Errorf("want %q, got %q", "fallback", got)
	}
}

func TestGetEnvOrFile_TrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret")
	if err := os.WriteFile(path, []byte("  padded-value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_SECRET_FILE", path)

	got, err := getEnvOrFile("TEST_SECRET", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "padded-value" {
		t.Errorf("want %q, got %q", "padded-value", got)
	}
}

func TestGetEnvOrFile_EmptyFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_SECRET_FILE", path)

	got, err := getEnvOrFile("TEST_SECRET", "default-not-used")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("want empty string, got %q", got)
	}
}
