// ======================================
// FILE: internal/config/config_test.go
// ======================================
package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create dummy config files
func createTempConfigFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	require.NoError(t, err)
}

func TestLoadConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir() // Create a temporary directory

	// Test loading with no config files present, relying on defaults
	cfg, err := config.LoadConfig(tmpDir)
	require.NoError(t, err)

	// Assert default values (examples)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 1*time.Hour, cfg.JWT.AccessTokenExpiry)
	assert.Equal(t, false, cfg.Minio.UseSSL)
	assert.Contains(t, cfg.Cors.AllowedOrigins, "http://localhost:3000") // Check default CORS
	assert.Equal(t, "default-insecure-secret-key-please-override", cfg.JWT.SecretKey) // Check default value

	// DSN should be empty if not set via env or file
	assert.Empty(t, cfg.Database.DSN)
}

func TestLoadConfig_BaseFile(t *testing.T) {
	tmpDir := t.TempDir()
	baseContent := `
server:
  port: "9090"
log:
  level: "warn"
database:
  dsn: "base_dsn_string"
jwt:
  secretKey: "base_secret"
`
	createTempConfigFile(t, tmpDir, "config.yaml", baseContent)

	cfg, err := config.LoadConfig(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "9090", cfg.Server.Port)        // Overridden by base
	assert.Equal(t, "warn", cfg.Log.Level)         // Overridden by base
	assert.Equal(t, "base_dsn_string", cfg.Database.DSN) // From base
	assert.Equal(t, "base_secret", cfg.JWT.SecretKey)   // From base
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)  // Default retained
}

func TestLoadConfig_EnvSpecificMerge(t *testing.T) {
	tmpDir := t.TempDir()
	baseContent := `
server:
  port: "8080" # Base port
log:
  level: "info"
database:
  dsn: "base_dsn"
minio:
  endpoint: "base_minio:9000"
`
	devContent := `
server:
  port: "8888" # Dev overrides port
log:
  level: "debug" # Dev overrides level
database:
  maxOpenConns: 10 # Dev overrides default
jwt:
  secretKey: "dev_secret" # Dev provides JWT secret
`
	createTempConfigFile(t, tmpDir, "config.yaml", baseContent)
	createTempConfigFile(t, tmpDir, "config.development.yaml", devContent)

	// Set environment for the test explicitly
	t.Setenv("APP_ENV", "development")

	cfg, err := config.LoadConfig(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "8888", cfg.Server.Port)          // Dev overrides base
	assert.Equal(t, "debug", cfg.Log.Level)          // Dev overrides base
	assert.Equal(t, "base_dsn", cfg.Database.DSN)    // From base (not in dev)
	assert.Equal(t, 10, cfg.Database.MaxOpenConns)    // Dev overrides default
	assert.Equal(t, "dev_secret", cfg.JWT.SecretKey)  // From dev
	assert.Equal(t, "base_minio:9000", cfg.Minio.Endpoint) // From base (not in dev)
}

func TestLoadConfig_EnvVariableOverride(t *testing.T) {
	tmpDir := t.TempDir()
	baseContent := `
server:
  port: "8080"
database:
  dsn: "file_dsn"
jwt:
  secretKey: "file_secret"
`
	createTempConfigFile(t, tmpDir, "config.yaml", baseContent)

	// Set environment variables
	t.Setenv("SERVER_PORT", "9999") // Overrides server.port
	t.Setenv("JWT_SECRETKEY", "env_secret") // Overrides jwt.secretKey
	t.Setenv("LOG_LEVEL", "error")     // Sets log.level (not in file)
	t.Setenv("DATABASE_DSN", "env_dsn") // Overrides database.dsn

	cfg, err := config.LoadConfig(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "9999", cfg.Server.Port)           // Overridden by env
	assert.Equal(t, "env_secret", cfg.JWT.SecretKey)   // Overridden by env
	assert.Equal(t, "error", cfg.Log.Level)          // Set by env
	assert.Equal(t, "env_dsn", cfg.Database.DSN)     // Overridden by env
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)   // Default retained
}

func TestLoadConfig_MissingRequiredDSN(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a config file *without* database.dsn
	content := `server: { port: "8080" }`
	createTempConfigFile(t, tmpDir, "config.yaml", content)

	// Ensure DATABASE_URL is not set
	t.Setenv("DATABASE_URL", "") // Explicitly unset for test isolation

	_, err := config.LoadConfig(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database DSN is required")
}

func TestLoadConfig_RequiredDSNFromEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a config file *without* database.dsn
	content := `server: { port: "8080" }`
	createTempConfigFile(t, tmpDir, "config.yaml", content)

	// Set DATABASE_URL environment variable
	expectedDSN := "env_dsn_for_fallback"
	t.Setenv("DATABASE_URL", expectedDSN)

	cfg, err := config.LoadConfig(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, expectedDSN, cfg.Database.DSN)
}