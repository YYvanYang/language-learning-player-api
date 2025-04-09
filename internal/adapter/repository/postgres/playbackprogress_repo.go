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

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

type PlaybackProgressRepository struct {
	db         *pgxpool.Pool
	logger     *slog.Logger
	getQuerier func(ctx context.Context) Querier
}

func NewPlaybackProgressRepository(db *pgxpool.Pool, logger *slog.Logger) *PlaybackProgressRepository {
	repo := &PlaybackProgressRepository{
		db:     db,
		logger: logger.With("repository", "PlaybackProgressRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db)
	}
	return repo
}

// --- Interface Implementation ---

func (r *PlaybackProgressRepository) Upsert(ctx context.Context, progress *domain.PlaybackProgress) error {
	q := r.getQuerier(ctx)
	progress.LastListenedAt = time.Now() // Update timestamp before saving
	query := `
        INSERT INTO playback_progress (user_id, track_id, progress_ms, last_listened_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, track_id) DO UPDATE SET
            progress_ms = EXCLUDED.progress_ms,
            last_listened_at = EXCLUDED.last_listened_at
    `
	_, err := q.Exec(ctx, query,
		progress.UserID,
		progress.TrackID,
		progress.Progress.Milliseconds(), // Point 1: Convert domain Duration to int64 ms
		progress.LastListenedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error upserting playback progress", "error", err, "userID", progress.UserID, "trackID", progress.TrackID)
		return fmt.Errorf("upserting playback progress: %w", err)
	}
	r.logger.DebugContext(ctx, "Playback progress upserted", "userID", progress.UserID, "trackID", progress.TrackID, "progressMs", progress.Progress.Milliseconds())
	return nil
}

func (r *PlaybackProgressRepository) Find(ctx context.Context, userID domain.UserID, trackID domain.TrackID) (*domain.PlaybackProgress, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT user_id, track_id, progress_ms, last_listened_at
        FROM playback_progress
        WHERE user_id = $1 AND track_id = $2
    `
	progress, err := r.scanProgress(ctx, q.QueryRow(ctx, query, userID, trackID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Error finding playback progress", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("finding playback progress: %w", err)
	}
	return progress, nil
}

func (r *PlaybackProgressRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.PlaybackProgress, int, error) {
	q := r.getQuerier(ctx)
	baseQuery := `FROM playback_progress WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT user_id, track_id, progress_ms, last_listened_at ` + baseQuery

	var total int
	err := q.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting progress by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting progress by user: %w", err)
	}
	if total == 0 {
		return []*domain.PlaybackProgress{}, 0, nil
	}

	orderByClause := " ORDER BY last_listened_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3)
	args := []interface{}{userID, page.Limit, page.Offset}
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing progress by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("listing progress by user: %w", err)
	}
	defer rows.Close()

	progressList := make([]*domain.PlaybackProgress, 0, page.Limit)
	for rows.Next() {
		progress, err := r.scanProgress(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning progress in ListByUser", "error", err)
			continue
		}
		progressList = append(progressList, progress)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating progress rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating progress rows: %w", err)
	}
	return progressList, total, nil
}

// Point 1: Updated scanProgress
func (r *PlaybackProgressRepository) scanProgress(ctx context.Context, row RowScanner) (*domain.PlaybackProgress, error) {
	var p domain.PlaybackProgress
	var progressMs int64 // Scan into int64

	err := row.Scan(
		&p.UserID,
		&p.TrackID,
		&progressMs, // Scan progress_ms column
		&p.LastListenedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert scanned milliseconds back to time.Duration
	p.Progress = time.Duration(progressMs) * time.Millisecond

	return &p, nil
}

var _ port.PlaybackProgressRepository = (*PlaybackProgressRepository)(nil)
