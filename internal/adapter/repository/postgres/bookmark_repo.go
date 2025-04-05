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
)

type BookmarkRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewBookmarkRepository(db *pgxpool.Pool, logger *slog.Logger) *BookmarkRepository {
	return &BookmarkRepository{
		db:     db,
		logger: logger.With("repository", "BookmarkRepository"),
	}
}

// --- Interface Implementation ---

func (r *BookmarkRepository) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	// CreatedAt is set by domain constructor or default value, ID by domain constructor
	query := `
        INSERT INTO bookmarks (id, user_id, track_id, timestamp_seconds, note, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.Exec(ctx, query,
		bookmark.ID,
		bookmark.UserID,
		bookmark.TrackID,
		bookmark.Timestamp.Seconds(), // Convert duration to seconds (float64, DB is INT)
		bookmark.Note,
		bookmark.CreatedAt,
	)
	if err != nil {
		// Check FK violations?
		r.logger.ErrorContext(ctx, "Error creating bookmark", "error", err, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
		return fmt.Errorf("creating bookmark: %w", err)
	}
	r.logger.InfoContext(ctx, "Bookmark created successfully", "bookmarkID", bookmark.ID, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
	return nil
}

func (r *BookmarkRepository) FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error) {
	query := `
        SELECT id, user_id, track_id, timestamp_seconds, note, created_at
        FROM bookmarks
        WHERE id = $1
    `
	bookmark, err := r.scanBookmark(ctx, r.db.QueryRow(ctx, query, id))
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
	query := `
        SELECT id, user_id, track_id, timestamp_seconds, note, created_at
        FROM bookmarks
        WHERE user_id = $1 AND track_id = $2
        ORDER BY timestamp_seconds ASC
    `
	rows, err := r.db.Query(ctx, query, userID, trackID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user and track", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("listing bookmarks by user and track: %w", err)
	}
	defer rows.Close()

	bookmarks := make([]*domain.Bookmark, 0)
	for rows.Next() {
		bookmark, err := r.scanBookmark(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUserAndTrack", "error", err)
			continue
		}
		bookmarks = append(bookmarks, bookmark)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUserAndTrack", "error", err)
		return nil, fmt.Errorf("iterating bookmark rows: %w", err)
	}

	return bookmarks, nil
}

func (r *BookmarkRepository) ListByUser(ctx context.Context, userID domain.UserID, page port.Page) ([]*domain.Bookmark, int, error) {
    baseQuery := `FROM bookmarks WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, user_id, track_id, timestamp_seconds, note, created_at ` + baseQuery

	// Count total
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting bookmarks by user: %w", err)
	}

	if total == 0 {
		return []*domain.Bookmark{}, 0, nil
	}

	// Build final query with sorting and pagination
	orderByClause := " ORDER BY created_at DESC" // Default sort by most recent
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3)
	args := []interface{}{userID, page.Limit, page.Offset}

	finalQuery := selectQuery + orderByClause + paginationClause

	// Execute query
	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("listing bookmarks by user: %w", err)
	}
	defer rows.Close()

	// Scan results
	bookmarks := make([]*domain.Bookmark, 0, page.Limit)
	for rows.Next() {
		bookmark, err := r.scanBookmark(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUser", "error", err)
			continue // Or return error
		}
		bookmarks = append(bookmarks, bookmark)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating bookmark rows: %w", err)
	}

	return bookmarks, total, nil
}


func (r *BookmarkRepository) Delete(ctx context.Context, id domain.BookmarkID) error {
    // Ownership check should happen in Usecase based on user from context vs bookmark's UserID
	query := `DELETE FROM bookmarks WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
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


// --- Helper Methods ---

// scanBookmark scans a single row into a domain.Bookmark.
func (r *BookmarkRepository) scanBookmark(ctx context.Context, row RowScanner) (*domain.Bookmark, error) {
	var b domain.Bookmark
	var timestampSec int // Scan into int first

	err := row.Scan(
		&b.ID,
		&b.UserID,
		&b.TrackID,
		&timestampSec, // Scan timestamp_seconds
		&b.Note,
		&b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert seconds back to duration
	b.Timestamp = time.Duration(timestampSec) * time.Second

	return &b, nil
}


// Compile-time check
var _ port.BookmarkRepository = (*BookmarkRepository)(nil)