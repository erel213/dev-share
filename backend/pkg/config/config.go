package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port           string `validate:"required"`
	BodyLimitBytes int    `validate:"gt=0"`

	// Database
	DBFilePath string `validate:"required"`

	// Auth
	JWTSecret      string `validate:"required,min=32"`
	AdminInitToken string

	// Encryption
	EncryptionKey []byte `validate:"required"`

	// Storage
	TemplateStoragePath string `validate:"required"`
	EnvExecutionPath    string `validate:"required"`

	// Terraform
	TFPluginCacheDir string

	// CORS
	CORSAllowOrigins string `validate:"required"`

	// Role-based secret access (valid values: "admin", "editor", "user")
	MinRoleViewSecrets string `validate:"required,oneof=admin editor user"`
	MinRoleEditSecrets string `validate:"required,oneof=admin editor user"`
}

// Load reads configuration from environment variables and returns a validated Config.
func Load() (*Config, error) {
	encryptionKeyHex, err := getEnvOrFile("ENCRYPTION_KEY", "")
	if err != nil {
		return nil, err
	}
	var encryptionKey []byte
	if encryptionKeyHex != "" {
		encryptionKey, err = hex.DecodeString(encryptionKeyHex)
		if err != nil {
			return nil, fmt.Errorf("ENCRYPTION_KEY must be a valid hex string: %w", err)
		}
	}

	bodyLimit, err := strconv.Atoi(getEnv("BODY_LIMIT_BYTES", "10485760"))
	if err != nil {
		return nil, fmt.Errorf("BODY_LIMIT_BYTES must be a valid integer: %w", err)
	}

	jwtSecret, err := getEnvOrFile("JWT_SECRET", "")
	if err != nil {
		return nil, err
	}
	adminInitToken, err := getEnvOrFile("ADMIN_INIT_TOKEN", "")
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		BodyLimitBytes:      bodyLimit,
		DBFilePath:          getEnv("DB_FILE_PATH", "./devshare.db"),
		JWTSecret:           jwtSecret,
		AdminInitToken:      adminInitToken,
		EncryptionKey:       encryptionKey,
		TemplateStoragePath: getEnv("TEMPLATE_STORAGE_PATH", "./template_storage"),
		EnvExecutionPath:    getEnv("ENV_EXECUTION_PATH", "./env_executions"),
		TFPluginCacheDir:    getEnv("TF_PLUGIN_CACHE_DIR", ""),
		CORSAllowOrigins:    getEnv("CORS_ALLOW_ORIGINS", "http://localhost:5173,http://localhost:3000"),
		MinRoleViewSecrets:  getEnv("MIN_ROLE_VIEW_SECRETS", "admin"),
		MinRoleEditSecrets:  getEnv("MIN_ROLE_EDIT_SECRETS", "admin"),
	}

	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvOrFile reads a secret from either ${key}_FILE (a file path, whose
// contents are read and whitespace-trimmed) or ${key} (a literal value). The
// _FILE form takes precedence when set — it's how Docker secrets, tmpfs-backed
// mounts, and the Postgres/MySQL/Vault ecosystems deliver secrets without
// exposing them via env vars visible to `docker inspect` or /proc/<pid>/environ.
func getEnvOrFile(key, defaultValue string) (string, error) {
	if path := os.Getenv(key + "_FILE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("%s_FILE: %w", key, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return getEnv(key, defaultValue), nil
}
