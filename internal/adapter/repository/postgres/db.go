// internal/adapter/repository/postgres/db.go
package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path
)

// NewPgxPool creates a new PostgreSQL connection pool.
func NewPgxPool(ctx context.Context, cfg config.DatabaseConfig, logger *slog.Logger) (*pgxpool.Pool, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	pgxConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	// Apply pool size settings from config
	pgxConfig.MaxConns = int32(cfg.MaxOpenConns)
	pgxConfig.MinConns = int32(cfg.MaxIdleConns) // pgx uses MinConns roughly like MaxIdleConns
	pgxConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	pgxConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	// Optional: Add logging hook for pgx
	// pgxConfig.ConnConfig.Logger = pgxLogger // Requires implementing pgx logging interface

	logger.Info("Connecting to PostgreSQL", "host", pgxConfig.ConnConfig.Host, "port", pgxConfig.ConnConfig.Port, "db", pgxConfig.ConnConfig.Database)

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err = pool.Ping(ctx); err != nil {
		pool.Close() // Close pool if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL")
	return pool, nil
}