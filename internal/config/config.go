// ============================================
// FILE: internal/config/config.go (MODIFIED)
// ============================================
package config

import (
	"fmt"
	"net/url"
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
	CDN      CDNConfig      `mapstructure:"cdn"`
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
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
}

// JWTConfig holds JWT related configuration.
type JWTConfig struct {
	SecretKey          string        `mapstructure:"secretKey"`
	AccessTokenExpiry  time.Duration `mapstructure:"accessTokenExpiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refreshTokenExpiry"` // ADDED
}

// MinioConfig holds MinIO connection configuration.
type MinioConfig struct {
	Endpoint        string        `mapstructure:"endpoint"`
	AccessKeyID     string        `mapstructure:"accessKeyId"`
	SecretAccessKey string        `mapstructure:"secretAccessKey"`
	UseSSL          bool          `mapstructure:"useSsl"`
	BucketName      string        `mapstructure:"bucketName"`
	PresignExpiry   time.Duration `mapstructure:"presignExpiry"`
}

// GoogleConfig holds Google OAuth configuration.
type GoogleConfig struct {
	ClientID     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"level"`
	JSON  bool   `mapstructure:"json"`
}

// CorsConfig holds CORS configuration.
type CorsConfig struct {
	AllowedOrigins   []string `mapstructure:"allowedOrigins"`
	AllowedMethods   []string `mapstructure:"allowedMethods"`
	AllowedHeaders   []string `mapstructure:"allowedHeaders"`
	AllowCredentials bool     `mapstructure:"allowCredentials"`
	MaxAge           int      `mapstructure:"maxAge"`
}

// CDNConfig holds optional CDN configuration.
type CDNConfig struct {
	BaseURL string `mapstructure:"baseUrl"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	v := viper.New()
	setDefaultValues(v) // Set defaults first

	v.AddConfigPath(path)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigName("config")
	if errReadBase := v.ReadInConfig(); errReadBase != nil {
		if _, ok := errReadBase.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("error reading base config file 'config.yaml': %w", errReadBase)
		}
		fmt.Fprintln(os.Stderr, "Info: Base config file 'config.yaml' not found. Skipping.")
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	if errReadEnv := v.MergeInConfig(); errReadEnv != nil {
		if _, ok := errReadEnv.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(os.Stderr, "Info: Environment config file 'config.%s.yaml' not found. Relying on base/defaults/env.\n", env)
		} else {
			return config, fmt.Errorf("error merging config file 'config.%s.yaml': %w", env, errReadEnv)
		}
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if config.Database.DSN == "" {
		envDSN := os.Getenv("DATABASE_URL")
		if envDSN != "" {
			fmt.Fprintln(os.Stderr, "Info: Database DSN not found in config, using DATABASE_URL environment variable.")
			config.Database.DSN = envDSN
		} else {
			return config, fmt.Errorf("database DSN is required but not found in config or DATABASE_URL environment variable")
		}
	}

	if config.CDN.BaseURL != "" {
		_, parseErr := url.ParseRequestURI(config.CDN.BaseURL)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Invalid CDN BaseURL configured ('%s'): %v. CDN rewriting will be disabled.\n", config.CDN.BaseURL, parseErr)
			config.CDN.BaseURL = ""
		}
	}

	// Validate token expirations
	if config.JWT.AccessTokenExpiry <= 0 {
		return config, fmt.Errorf("jwt.accessTokenExpiry must be a positive duration")
	}
	if config.JWT.RefreshTokenExpiry <= 0 {
		return config, fmt.Errorf("jwt.refreshTokenExpiry must be a positive duration")
	}
	if config.JWT.RefreshTokenExpiry <= config.JWT.AccessTokenExpiry {
		return config, fmt.Errorf("jwt.refreshTokenExpiry must be longer than jwt.accessTokenExpiry")
	}

	return config, nil
}

func setDefaultValues(v *viper.Viper) {
	// Server Defaults
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.readTimeout", "5s")
	v.SetDefault("server.writeTimeout", "10s")
	v.SetDefault("server.idleTimeout", "120s")

	// Database Defaults
	v.SetDefault("database.maxOpenConns", 25)
	v.SetDefault("database.maxIdleConns", 25)
	v.SetDefault("database.connMaxLifetime", "5m")
	v.SetDefault("database.connMaxIdleTime", "5m")

	// JWT Defaults
	v.SetDefault("jwt.secretKey", "default-insecure-secret-key-please-override")
	v.SetDefault("jwt.accessTokenExpiry", "1h")
	v.SetDefault("jwt.refreshTokenExpiry", "720h") // Default: 30 days (ADDED)

	// MinIO Defaults
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.accessKeyId", "minioadmin")
	v.SetDefault("minio.secretAccessKey", "minioadmin")
	v.SetDefault("minio.useSsl", false)
	v.SetDefault("minio.bucketName", "language-audio")
	v.SetDefault("minio.presignExpiry", "1h")

	// Google Defaults
	v.SetDefault("google.clientId", "")
	v.SetDefault("google.clientSecret", "")

	// Log Defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.json", false)

	// CORS Defaults
	v.SetDefault("cors.allowedOrigins", []string{"http://localhost:3000", "http://127.0.0.1:3000"})
	v.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowedHeaders", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"})
	v.SetDefault("cors.allowCredentials", true)
	v.SetDefault("cors.maxAge", 300)

	// CDN Default
	v.SetDefault("cdn.baseUrl", "")
}

func GetConfig() (Config, error) {
	return LoadConfig(".")
}
