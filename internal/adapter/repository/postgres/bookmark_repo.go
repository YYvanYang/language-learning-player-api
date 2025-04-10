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

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// BookmarkRepository implements port.BookmarkRepository using PostgreSQL.
type BookmarkRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// getQuerier retrieves the correct Querier (Pool or Tx) from context.
	getQuerier func(ctx context.Context) Querier
}

// NewBookmarkRepository creates a new BookmarkRepository.
func NewBookmarkRepository(db *pgxpool.Pool, logger *slog.Logger) *BookmarkRepository {
	repo := &BookmarkRepository{
		db:     db,
		logger: logger.With("repository", "BookmarkRepository"),
	}
	// Initialize the helper function to get the Querier from context or pool.
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Uses the helper from tx_manager.go
	}
	return repo
}

// --- Interface Implementation ---

// Create inserts a new bookmark record into the database.
func (r *BookmarkRepository) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	q := r.getQuerier(ctx) // Get Pool or Transaction Querier

	// Ensure CreatedAt is set (usually done by domain constructor, but good fallback)
	if bookmark.CreatedAt.IsZero() {
		bookmark.CreatedAt = time.Now()
	}
	// Ensure ID is set (should be done by domain constructor)
	if bookmark.ID == (domain.BookmarkID{}) {
		bookmark.ID = domain.NewBookmarkID()
	}

	// SQL uses timestamp_ms column (BIGINT)
	query := `
        INSERT INTO bookmarks (id, user_id, track_id, timestamp_ms, note, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := q.Exec(ctx, query,
		bookmark.ID,
		bookmark.UserID,
		bookmark.TrackID,
		bookmark.Timestamp, // Use time.Duration directly, pgx handles INTERVAL
		bookmark.Note,
		bookmark.CreatedAt,
	)

	if err != nil {
		// Consider checking for FK violation on user_id or track_id
		r.logger.ErrorContext(ctx, "Error creating bookmark", "error", err, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
		return fmt.Errorf("creating bookmark: %w", err)
	}

	r.logger.InfoContext(ctx, "Bookmark created successfully", "bookmarkID", bookmark.ID, "userID", bookmark.UserID, "trackID", bookmark.TrackID)
	return nil
}

// FindByID retrieves a bookmark by its unique ID.
func (r *BookmarkRepository) FindByID(ctx context.Context, id domain.BookmarkID) (*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	// SQL selects timestamp_ms column
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at
        FROM bookmarks
        WHERE id = $1
    `
	// Use the scanBookmark helper which handles the ms -> duration conversion
	bookmark, err := r.scanBookmark(ctx, q.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map DB error to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding bookmark by ID", "error", err, "bookmarkID", id)
		return nil, fmt.Errorf("finding bookmark by ID: %w", err)
	}
	return bookmark, nil
}

// ListByUserAndTrack retrieves all bookmarks for a specific user on a specific track, ordered by timestamp.
func (r *BookmarkRepository) ListByUserAndTrack(ctx context.Context, userID domain.UserID, trackID domain.TrackID) ([]*domain.Bookmark, error) {
	q := r.getQuerier(ctx)
	// SQL selects timestamp_ms and orders by it
	query := `
        SELECT id, user_id, track_id, timestamp_ms, note, created_at
        FROM bookmarks
        WHERE user_id = $1 AND track_id = $2
        ORDER BY timestamp_ms ASC
    `
	rows, err := q.Query(ctx, query, userID, trackID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user and track", "error", err, "userID", userID, "trackID", trackID)
		return nil, fmt.Errorf("listing bookmarks by user and track: %w", err)
	}
	defer rows.Close()

	bookmarks := make([]*domain.Bookmark, 0)
	for rows.Next() {
		// Use the scanBookmark helper
		bookmark, err := r.scanBookmark(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUserAndTrack", "error", err)
			continue // Skip faulty row? Or return error? Let's skip for now.
		}
		bookmarks = append(bookmarks, bookmark)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUserAndTrack", "error", err)
		return nil, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, nil
}

// ListByUser retrieves a paginated list of all bookmarks for a user, ordered by creation time descending.
func (r *BookmarkRepository) ListByUser(ctx context.Context, userID domain.UserID, page pagination.Page) ([]*domain.Bookmark, int, error) {
	q := r.getQuerier(ctx)
	baseQuery := `FROM bookmarks WHERE user_id = $1`
	countQuery := `SELECT count(*) ` + baseQuery
	// SQL selects timestamp_ms column
	selectQuery := `SELECT id, user_id, track_id, timestamp_ms, note, created_at ` + baseQuery

	var total int
	err := q.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting bookmarks by user", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("counting bookmarks by user: %w", err)
	}
	if total == 0 {
		return []*domain.Bookmark{}, 0, nil
	}

	// Use page.Limit and page.Offset directly as they are validated/defaulted by the pagination package
	orderByClause := " ORDER BY created_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", 2, 3) // Arguments start from $2
	args := []interface{}{userID, page.Limit, page.Offset}
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing bookmarks by user", "error", err, "userID", userID, "page", page)
		return nil, 0, fmt.Errorf("listing bookmarks by user: %w", err)
	}
	defer rows.Close()

	bookmarks := make([]*domain.Bookmark, 0, page.Limit)
	for rows.Next() {
		// Use the scanBookmark helper
		bookmark, err := r.scanBookmark(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning bookmark in ListByUser", "error", err)
			continue // Skip faulty row
		}
		bookmarks = append(bookmarks, bookmark)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating bookmark rows in ListByUser", "error", err)
		return nil, 0, fmt.Errorf("iterating bookmark rows: %w", err)
	}
	return bookmarks, total, nil
}

// Delete removes a bookmark by its ID. Ownership check is expected to be done in the Usecase layer before calling this.
func (r *BookmarkRepository) Delete(ctx context.Context, id domain.BookmarkID) error {
	q := r.getQuerier(ctx)
	query := `DELETE FROM bookmarks WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting bookmark", "error", err, "bookmarkID", id)
		return fmt.Errorf("deleting bookmark: %w", err)
	}
	// Check if any row was actually deleted
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound // Bookmark ID did not exist
	}
	r.logger.InfoContext(ctx, "Bookmark deleted successfully", "bookmarkID", id)
	return nil
}

// scanBookmark is a helper function to scan a single row into a domain.Bookmark.
// It handles the conversion from the database's timestamp_ms (BIGINT) to domain's Timestamp (time.Duration).
func (r *BookmarkRepository) scanBookmark(ctx context.Context, row RowScanner) (*domain.Bookmark, error) {
	var b domain.Bookmark
	// Scan directly into time.Duration field

	err := row.Scan(
		&b.ID,
		&b.UserID,
		&b.TrackID,
		&b.Timestamp, // Scan INTERVAL directly into time.Duration
		&b.Note,
		&b.CreatedAt,
	)
	if err != nil {
		return nil, err // Propagate scan errors (including pgx.ErrNoRows)
	}

	return &b, nil
}

// Compile-time check to ensure BookmarkRepository satisfies the port.BookmarkRepository interface.
var _ port.BookmarkRepository = (*BookmarkRepository)(nil)
