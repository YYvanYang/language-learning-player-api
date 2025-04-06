// internal/adapter/repository/postgres/bookmark_repo.go
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
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import pagination
)

type BookmarkRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// Use Querier interface to work with both pool and transaction
	getQuerier func(ctx context.Context) Querier
}

func NewBookmarkRepository(db *pgxpool.Pool, logger *slog.Logger) *BookmarkRepository {
	repo := &BookmarkRepository{
		db:     db,
		logger: logger.With("repository", "BookmarkRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Use the helper
	}
	return repo
}

// --- Interface Implementation ---

func (r *BookmarkRepository) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	q := r.getQuerier(ctx)
	query := `
        INSERT INTO bookmarks (id, user_id, track_id, timestamp_ms, note, created_at) -- CHANGED column name
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := q.Exec(ctx, query,
		bookmark.ID,
		bookmark.UserID,
		bookmark.TrackID,
		bookmark.Timestamp.Milliseconds(), // CORRECTED: Save milliseconds
		bookmark.Note,
		bookmark.CreatedAt,
	)
	if err != nil {
		// Consider checking FK violations here
		r.logger.ErrorContext(ctx, "Error creating bookmark", "error", err, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
		return fmt.Errorf("creating bookmark: %w", err)
	}
	r.logger.InfoContext(ctx, "Bookmark created successfully", "bookmarkID", bookmark.ID, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
	return nil
}

func (r *BookmarkRepository) FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at -- CHANGED column name
        FROM bookmarks
        WHERE id = $1
    `
	bookmark, err := r.scanBookmark(ctx, q.QueryRow(ctx, query, id)) // Pass QueryRow
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Error finding bookmark by ID", "error", err, "bookmarkID", id)
		return nil, fmt.Errorf("finding bookmark by ID: %w", err)
	}
	return bookmark, nil
}

func (r *BookmarkRepository) ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at -- CHANGED column name
        FROM bookmarks
        WHERE user_id = $1 AND track_id = $2
        ORDER BY timestamp_ms ASC -- CHANGED column name
    `
	rows, err := q.Query(ctx, query, userID, trackID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user and track", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("listing bookmarks by user and track: %w", err)
	}
	defer rows.Close()
	bookmarks := make([]*domain.Bookmark, 0)
	for rows.Next() {
		bookmark, err := r.scanBookmark(ctx, rows) // Use RowScanner compatible scan
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUserAndTrack", "error", err); continue }
		bookmarks = append(bookmarks, bookmark)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUserAndTrack", "error", err)
		return nil, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, nil
}

func (r *BookmarkRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.Bookmark, int, error) {
	q := r.getQuerier(ctx)
	baseQuery := `FROM bookmarks WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, user_id, track_id, timestamp_ms, note, created_at ` + baseQuery // CHANGED column name

	var total int
	err := q.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting bookmarks by user: %w", err)
	}
	if total == 0 { return []*domain.Bookmark{}, 0, nil }

	orderByClause := " ORDER BY created_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3)
	args := []interface{}{userID, page.Limit, page.Offset}
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("listing bookmarks by user: %w", err)
	}
	defer rows.Close()
	bookmarks := make([]*domain.Bookmark, 0, page.Limit)
	for rows.Next() {
		bookmark, err := r.scanBookmark(ctx, rows) // Use RowScanner compatible scan
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUser", "error", err); continue }
		bookmarks = append(bookmarks, bookmark)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, total, nil
}

func (r *BookmarkRepository) Delete(ctx context.Context, id domain.BookmarkID) error {
	// Ownership check done in Usecase
	q := r.getQuerier(ctx)
	query := `DELETE FROM bookmarks WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting bookmark", "error", err, "bookmarkID", id)
		return fmt.Errorf("deleting bookmark: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	r.logger.InfoContext(ctx, "Bookmark deleted successfully", "bookmarkID", id)
	return nil
}

// scanBookmark scans a single row into a domain.Bookmark.
// Accepts RowScanner interface
func (r *BookmarkRepository) scanBookmark(ctx context.Context, row RowScanner) (*domain.Bookmark, error) {
	var b domain.Bookmark
	var timestampMs int64 // CORRECTED: Scan milliseconds into int64

	err := row.Scan(
		&b.ID,
		&b.UserID,
		&b.TrackID,
		&timestampMs, // **FIXED:** Scan timestamp_ms (Changed '×' to 't')
		&b.Note,
		&b.CreatedAt,
	)
	if err != nil { return nil, err }

	// CORRECTED: Convert milliseconds back to duration
	b.Timestamp = time.Duration(timestampMs) * time.Millisecond

	return &b, nil
}

var _ port.BookmarkRepository = (*BookmarkRepository)(nil)