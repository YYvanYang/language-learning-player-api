// pkg/logger/logger.go
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/yvanyang/language-learning-player-api/internal/config" // Adjust import path
)

// NewLogger creates and returns a new slog.Logger based on the provided configuration.
func NewLogger(cfg config.LogConfig) *slog.Logger {
	var logWriter io.Writer = os.Stdout // Default to standard output

	// Determine log level
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo // Default to Info level if unspecified or invalid
	}

	opts := &slog.HandlerOptions{
		Level: level,
		// AddSource: true, // Uncomment this to include source file and line number in logs (can affect performance)
	}

	var handler slog.Handler
	if cfg.JSON {
		handler = slog.NewJSONHandler(logWriter, opts)
	} else {
		handler = slog.NewTextHandler(logWriter, opts)
	}

	logger := slog.New(handler)
	return logger
}
