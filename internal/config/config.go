// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Minio    MinioConfig    `mapstructure:"minio"`
	Google   GoogleConfig   `mapstructure:"google"`
	Log      LogConfig      `mapstructure:"log"`
	Cors     CorsConfig     `mapstructure:"cors"`
}

// ServerConfig holds server specific configuration.
type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"` // Data Source Name (e.g., postgresql://user:password@host:port/dbname?sslmode=disable)
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
}

// JWTConfig holds JWT related configuration.
type JWTConfig struct {
	SecretKey         string        `mapstructure:"secretKey"`
	AccessTokenExpiry time.Duration `mapstructure:"accessTokenExpiry"`
	// RefreshTokenExpiry time.Duration `mapstructure:"refreshTokenExpiry"` // Add if implementing refresh tokens
}

// MinioConfig holds MinIO connection configuration.
type MinioConfig struct {
	Endpoint        string        `mapstructure:"endpoint"`
	AccessKeyID     string        `mapstructure:"accessKeyId"`
	SecretAccessKey string        `mapstructure:"secretAccessKey"`
	UseSSL          bool          `mapstructure:"useSsl"`
	BucketName      string        `mapstructure:"bucketName"`
	PresignExpiry   time.Duration `mapstructure:"presignExpiry"` // Default expiry for presigned URLs
}

// GoogleConfig holds Google OAuth configuration.
type GoogleConfig struct {
	ClientID     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
	// RedirectURL string `mapstructure:"redirectUrl"` // Usually handled by frontend, but maybe needed later
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"` // e.g., "debug", "info", "warn", "error"
	JSON  bool   `mapstructure:"json"`  // Output logs in JSON format
}

// CorsConfig holds CORS configuration.
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowedOrigins"`
	AllowedMethods   []string `mapstructure:"allowedMethods"`
	AllowedHeaders   []string `mapstructure:"allowedHeaders"`
	AllowCredentials bool     `mapstructure:"allowCredentials"`
	MaxAge           int      `mapstructure:"maxAge"`
}


// LoadConfig reads configuration from file or environment variables.
// It looks for config files in the specified path (e.g., "." for current directory).
func LoadConfig(path string) (config Config, err error) {
	v := viper.New()

	// Set default values first
	setDefaultValues(v)

	// Configure Viper
	v.AddConfigPath(path)         // Path to look for the config file in
	v.SetConfigType("yaml")
	v.AutomaticEnv()              // Read in environment variables that match
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores for env var names (e.g., server.port -> SERVER_PORT)

	// 1. Load base configuration (config.yaml)
	v.SetConfigName("config")
	if errReadBase := v.ReadInConfig(); errReadBase != nil {
		if _, ok := errReadBase.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a 'file not found' error for the base file
			return config, fmt.Errorf("error reading base config file 'config.yaml': %w", errReadBase)
		}
		// Base file not found is okay, we can proceed without it.
		fmt.Fprintln(os.Stderr, "Info: Base config file 'config.yaml' not found. Skipping.")
	}

	// 2. Load and merge environment-specific configuration (e.g., config.development.yaml)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development" // Default environment
	}
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	if errReadEnv := v.MergeInConfig(); errReadEnv != nil {
		if _, ok := errReadEnv.(viper.ConfigFileNotFoundError); ok {
			// Env specific file not found, rely on base/env vars/defaults
			fmt.Fprintf(os.Stderr, "Info: Environment config file 'config.%s.yaml' not found. Relying on base/defaults/env.\n", env)
		} else {
			// Other error reading env config file
			return config, fmt.Errorf("error merging config file 'config.%s.yaml': %w", env, errReadEnv)
		}
	}

	// Unmarshal the final merged configuration
	err = v.Unmarshal(&config)
	if err != nil {
		 return config, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Optional: Add validation logic here if needed
    // For example, check if required fields like database DSN are set:
    if config.Database.DSN == "" {
        // Attempt to get from DATABASE_URL env var as a fallback
        envDSN := os.Getenv("DATABASE_URL")
        if envDSN != "" {
            fmt.Fprintln(os.Stderr, "Info: Database DSN not found in config, using DATABASE_URL environment variable.")
            config.Database.DSN = envDSN
        } else {
            // If still empty after checking env var, return error
            return config, fmt.Errorf("database DSN is required but not found in config or DATABASE_URL environment variable")
        }
    }

	return config, nil
}

// Use a dedicated viper instance for defaults to avoid conflicts
func setDefaultValues(v *viper.Viper) {
    // Server Defaults
    v.SetDefault("server.port", "8080")
    v.SetDefault("server.readTimeout", "5s")
    v.SetDefault("server.writeTimeout", "10s")
    v.SetDefault("server.idleTimeout", "120s")

    // Database Defaults (Note: DSN has no sensible default)
    v.SetDefault("database.maxOpenConns", 25)
    v.SetDefault("database.maxIdleConns", 25)
    v.SetDefault("database.connMaxLifetime", "5m")
	v.SetDefault("database.connMaxIdleTime", "5m") // Usually same as lifetime

    // JWT Defaults
	v.SetDefault("jwt.secretKey", "default-insecure-secret-key-please-override") // Provide a default but stress it's insecure
    v.SetDefault("jwt.accessTokenExpiry", "1h")

    // MinIO Defaults
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.accessKeyId", "minioadmin")
	v.SetDefault("minio.secretAccessKey", "minioadmin")
    v.SetDefault("minio.useSsl", false)
    v.SetDefault("minio.presignExpiry", "1h") // Default URL expiry

	// Google Defaults
	v.SetDefault("google.clientId", "")
	v.SetDefault("google.clientSecret", "")

    // Log Defaults
    v.SetDefault("log.level", "info")
    v.SetDefault("log.json", false) // Easier to read logs for dev default

	// CORS Defaults (Example: Allow local dev server)
	v.SetDefault("cors.allowedOrigins", []string{"http://localhost:3000", "http://127.0.0.1:3000"})
	v.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowedHeaders", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"})
	v.SetDefault("cors.allowCredentials", true)
	v.SetDefault("cors.maxAge", 300) // 5 minutes
}

// GetConfig is a helper function to load config, often called from main.
// Assumes config files are in the current working directory (".")
func GetConfig() (Config, error) {
    return LoadConfig(".")
}