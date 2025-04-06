// internal/adapter/repository/postgres/playbackprogress_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"                       // Import pagination
)

type PlaybackProgressRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewPlaybackProgressRepository(db *pgxpool.Pool, logger *slog.Logger) *PlaybackProgressRepository {
	return &PlaybackProgressRepository{
		db:     db,
		logger: logger.With("repository", "PlaybackProgressRepository"),
	}
}

// --- Interface Implementation ---

func (r *PlaybackProgressRepository) Upsert(ctx context.Context, progress *domain.PlaybackProgress) error {
	// Ensure LastListenedAt is current before upserting
	progress.LastListenedAt = time.Now()

	query := `
        INSERT INTO playback_progress (user_id, track_id, progress_seconds, last_listened_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, track_id) DO UPDATE SET
            progress_seconds = EXCLUDED.progress_seconds,
            last_listened_at = EXCLUDED.last_listened_at
    `
	_, err := r.db.Exec(ctx, query,
		progress.UserID,
		progress.TrackID,
		progress.Progress.Seconds(), // Convert duration to seconds (float64, but DB column is INT)
		progress.LastListenedAt,
	)

	if err != nil {
		// Check for foreign key violations (user_id, track_id)
		// var pgErr *pgconn.PgError
		// if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation { ... return domain.ErrNotFound? or specific error? }
		r.logger.ErrorContext(ctx, "Error upserting playback progress", "error", err, "userID", progress.UserID, "trackID", progress.TrackID)
		return fmt.Errorf("upserting playback progress: %w", err)
	}

	r.logger.DebugContext(ctx, "Playback progress upserted", "userID", progress.UserID, "trackID", progress.TrackID, "progressSec", progress.Progress.Seconds())
	return nil
}


func (r *PlaybackProgressRepository) Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	query := `
        SELECT user_id, track_id, progress_seconds, last_listened_at
        FROM playback_progress
        WHERE user_id = $1 AND track_id = $2
    `
	progress, err := r.scanProgress(ctx, r.db.QueryRow(ctx, query, userID, trackID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Not found is expected, map it
		}
		r.logger.ErrorContext(ctx, "Error finding playback progress", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("finding playback progress: %w", err)
	}
	return progress, nil
}

func (r *PlaybackProgressRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.PlaybackProgress, int, error) {
    baseQuery := `FROM playback_progress WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT user_id, track_id, progress_seconds, last_listened_at ` + baseQuery

	// Count total
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting progress by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting progress by user: %w", err)
	}

	if total == 0 {
		return []*domain.PlaybackProgress{}, 0, nil
	}

	// Build final query with sorting and pagination
	orderByClause := " ORDER BY last_listened_at DESC" // Default sort by most recent
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3)
	args := []interface{}{userID, page.Limit, page.Offset}

	finalQuery := selectQuery + orderByClause + paginationClause

	// Execute query
	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing progress by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("listing progress by user: %w", err)
	}
	defer rows.Close()

	// Scan results
	progressList := make([]*domain.PlaybackProgress, 0, page.Limit)
	for rows.Next() {
		progress, err := r.scanProgress(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning progress in ListByUser", "error", err)
			continue // Or return error
		}
		progressList = append(progressList, progress)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating progress rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating progress rows: %w", err)
	}

	return progressList, total, nil
}

// --- Helper Methods ---

// scanProgress scans a single row into a domain.PlaybackProgress.
func (r *PlaybackProgressRepository) scanProgress(ctx context.Context, row RowScanner) (*domain.PlaybackProgress, error) {
	var p domain.PlaybackProgress
	var progressSec int // Scan seconds into int

	err := row.Scan(
		&p.UserID,
		&p.TrackID,
		&progressSec, // Scan progress_seconds
		&p.LastListenedAt,
	)
	if err != nil {
		return nil, err // Let caller handle pgx.ErrNoRows or other DB errors
	}

	// Convert seconds back to duration
	p.Progress = time.Duration(progressSec) * time.Second

	return &p, nil
}

// Compile-time check
var _ port.PlaybackProgressRepository = (*PlaybackProgressRepository)(nil)