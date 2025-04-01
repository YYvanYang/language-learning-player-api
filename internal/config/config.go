// internal/config/config.go
package config

import (
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
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyId"`
	SecretAccessKey string `mapstructure:"secretAccessKey"`
	UseSSL          bool   `mapstructure:"useSsl"`
	BucketName      string `mapstructure:"bucketName"`
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
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)         // Path to look for the config file in
	viper.SetConfigName("config")     // Name of config file (without extension)
	viper.SetConfigType("yaml")       // REQUIRED if the config file does not have the extension in the name

	viper.AutomaticEnv()              // Read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores for env var names (e.g., server.port -> SERVER_PORT)

	// Set default values (optional but recommended)
	setDefaultValues()

	err = viper.ReadInConfig()        // Find and read the config file
	if err != nil {
		// If config file not found, continue (might rely solely on env vars or defaults)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return config, err
		}
		// If file not found, that's okay if env vars are set or defaults are sufficient.
		// We can add a log here later if needed.
	}

	err = viper.Unmarshal(&config)    // Unmarshal config into struct
	return config, err
}

func setDefaultValues() {
    // Server Defaults
    viper.SetDefault("server.port", "8080")
    viper.SetDefault("server.readTimeout", "5s")
    viper.SetDefault("server.writeTimeout", "10s")
    viper.SetDefault("server.idleTimeout", "120s")

    // Database Defaults
    viper.SetDefault("database.maxOpenConns", 25)
    viper.SetDefault("database.maxIdleConns", 25)
    viper.SetDefault("database.connMaxLifetime", "5m")
	viper.SetDefault("database.connMaxIdleTime", "5m") // Usually same as lifetime

    // JWT Defaults
    viper.SetDefault("jwt.accessTokenExpiry", "1h")
    // viper.SetDefault("jwt.refreshTokenExpiry", "720h") // ~30 days

    // MinIO Defaults
    viper.SetDefault("minio.useSsl", false)
    viper.SetDefault("minio.presignExpiry", "1h") // Default URL expiry

    // Log Defaults
    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.json", true) // Prefer JSON logs for production

	// CORS Defaults (Example: Allow all for local dev, restrict in prod)
	viper.SetDefault("cors.allowedOrigins", []string{"*"}) // Be careful with "*" in production!
	viper.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowedHeaders", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"})
	viper.SetDefault("cors.allowCredentials", true)
	viper.SetDefault("cors.maxAge", 300) // 5 minutes
}