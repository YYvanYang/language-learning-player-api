// internal/adapter/repository/postgres/audiotrack_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq" // Using lib/pq for array handling with pgx (can be tricky otherwise)

	"your_project/internal/domain" // Adjust import path
	"your_project/internal/port"   // Adjust import path
)

type AudioTrackRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewAudioTrackRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioTrackRepository {
	return &AudioTrackRepository{
		db:     db,
		logger: logger.With("repository", "AudioTrackRepository"),
	}
}

// --- Interface Implementation ---

func (r *AudioTrackRepository) Create(ctx context.Context, track *domain.AudioTrack) error {
	query := `
		INSERT INTO audio_tracks
			(id, title, description, language_code, level, duration_ms,
			 minio_bucket, minio_object_key, cover_image_url, uploader_id,
			 is_public, tags, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(ctx, query,
		track.ID,
		track.Title,
		track.Description,
		track.Language.Code(), // Use code from Language VO
		track.Level,           // Use AudioLevel directly (string)
		track.Duration.Milliseconds(), // Store as milliseconds integer
		track.MinioBucket,
		track.MinioObjectKey,
		track.CoverImageURL,
		track.UploaderID,
		track.IsPublic,
		pq.Array(track.Tags), // Use pq.Array to handle []string -> TEXT[]
		track.CreatedAt,
		track.UpdatedAt,
	)

	if err != nil {
		// TODO: Check for unique constraint violation (minio_object_key)
		r.logger.ErrorContext(ctx, "Error creating audio track", "error", err, "trackID", track.ID)
		return fmt.Errorf("creating audio track: %w", err)
	}
	r.logger.InfoContext(ctx, "Audio track created successfully", "trackID", track.ID, "title", track.Title)
	return nil
}

func (r *AudioTrackRepository) FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error) {
	query := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = $1
    `
	track, err := r.scanTrack(ctx, r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding audio track by ID", "error", err, "trackID", id)
		return nil, fmt.Errorf("finding audio track by ID: %w", err)
	}
	return track, nil
}

func (r *AudioTrackRepository) ListByIDs(ctx context.Context, ids []domain.TrackID) ([]*domain.AudioTrack, error) {
    if len(ids) == 0 {
        return []*domain.AudioTrack{}, nil
    }

	// Convert domain.TrackID slice to primitive UUID slice for the query parameter
    uuidStrs := make([]string, len(ids))
    for i, id := range ids {
        uuidStrs[i] = id.String()
    }

    query := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = ANY($1) -- Use = ANY() for array matching
		ORDER BY array_position($1, id::text) -- Attempt to preserve order (requires casting id to text)
    `
	// Note: array_position might not be the most performant way for large lists.
	// An alternative is fetching all and re-ordering in Go code based on the input `ids`.

	rows, err := r.db.Query(ctx, query, uuidStrs)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks by IDs", "error", err)
		return nil, fmt.Errorf("listing audio tracks by IDs: %w", err)
	}
	defer rows.Close()

	tracks := make([]*domain.AudioTrack, 0, len(ids))
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows) // Use pgx.Rows which satisfies the RowScanner interface
		if err != nil {
			// Log error but continue processing other rows if possible
			r.logger.ErrorContext(ctx, "Error scanning track in ListByIDs", "error", err)
			continue // Or return error immediately depending on desired behavior
		}
		tracks = append(tracks, track)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track rows in ListByIDs", "error", err)
		return nil, fmt.Errorf("iterating track rows: %w", err)
	}

	// Re-ordering in Go code (alternative to ORDER BY array_position)
	// trackMap := make(map[domain.TrackID]*domain.AudioTrack, len(tracks))
	// for _, t := range tracks {
	// 	trackMap[t.ID] = t
	// }
	// orderedTracks := make([]*domain.AudioTrack, 0, len(ids))
	// for _, id := range ids {
	// 	if t, ok := trackMap[id]; ok {
	// 		orderedTracks = append(orderedTracks, t)
	// 	}
	// }
	// return orderedTracks, nil

	return tracks, nil
}


func (r *AudioTrackRepository) List(ctx context.Context, params port.ListTracksParams, page port.Page) ([]*domain.AudioTrack, int, error) {
    var args []interface{}
    argID := 1

	// Base query
	baseQuery := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
    `
	countQuery := `SELECT count(*) FROM audio_tracks`
	whereClause := " WHERE 1=1" // Start with true condition

	// Build WHERE clause dynamically
	if params.Query != nil && *params.Query != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argID, argID) // Case-insensitive search
		args = append(args, "%"+*params.Query+"%")
		argID++
	}
	if params.LanguageCode != nil && *params.LanguageCode != "" {
		whereClause += fmt.Sprintf(" AND language_code = $%d", argID)
		args = append(args, *params.LanguageCode)
		argID++
	}
	if params.Level != nil && *params.Level != "" {
		whereClause += fmt.Sprintf(" AND level = $%d", argID)
		args = append(args, *params.Level)
		argID++
	}
	if params.IsPublic != nil {
		whereClause += fmt.Sprintf(" AND is_public = $%d", argID)
		args = append(args, *params.IsPublic)
		argID++
	}
	if params.UploaderID != nil {
		whereClause += fmt.Sprintf(" AND uploader_id = $%d", argID)
		args = append(args, *params.UploaderID)
		argID++
	}
	if len(params.Tags) > 0 {
		whereClause += fmt.Sprintf(" AND tags @> $%d", argID) // Check if array contains elements
		args = append(args, pq.Array(params.Tags))
		argID++
	}

	// Get total count matching filters
	var total int
	countQuery += whereClause
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting audio tracks", "error", err, "filters", params)
		return nil, 0, fmt.Errorf("counting audio tracks: %w", err)
	}

	if total == 0 {
		return []*domain.AudioTrack{}, 0, nil
	}


	// Build final query with sorting and pagination
	orderByClause := " ORDER BY created_at DESC" // Default sort
	if params.SortBy != "" {
		// Validate SortBy to prevent SQL injection
		allowedSorts := map[string]string{
			"createdAt": "created_at",
			"title":     "title",
			"duration":  "duration_ms",
			"level":     "level", // Add more allowed fields
		}
		dbColumn, ok := allowedSorts[params.SortBy]
		if ok {
			direction := " ASC"
			if strings.ToLower(params.SortDirection) == "desc" {
				direction = " DESC"
			}
			orderByClause = fmt.Sprintf(" ORDER BY %s%s", dbColumn, direction)
		} else {
			r.logger.WarnContext(ctx, "Invalid sort field requested", "sortBy", params.SortBy)
			// Keep default sort
		}
	}

	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)

	finalQuery := baseQuery + whereClause + orderByClause + paginationClause

	// Execute query
	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks", "error", err, "filters", params, "page", page)
		return nil, 0, fmt.Errorf("listing audio tracks: %w", err)
	}
	defer rows.Close()

	// Scan results
	tracks := make([]*domain.AudioTrack, 0, page.Limit)
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning track in List", "error", err)
			continue // Or return error
		}
		tracks = append(tracks, track)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track rows in List", "error", err)
		return nil, 0, fmt.Errorf("iterating track rows: %w", err)
	}


	return tracks, total, nil
}


func (r *AudioTrackRepository) Update(ctx context.Context, track *domain.AudioTrack) error {
	track.UpdatedAt = time.Now()
	query := `
		UPDATE audio_tracks SET
			title = $2, description = $3, language_code = $4, level = $5, duration_ms = $6,
			minio_bucket = $7, minio_object_key = $8, cover_image_url = $9, uploader_id = $10,
			is_public = $11, tags = $12, updated_at = $13
		WHERE id = $1
	`
	cmdTag, err := r.db.Exec(ctx, query,
		track.ID,
		track.Title,
		track.Description,
		track.Language.Code(),
		track.Level,
		track.Duration.Milliseconds(),
		track.MinioBucket,
		track.MinioObjectKey,
		track.CoverImageURL,
		track.UploaderID,
		track.IsPublic,
		pq.Array(track.Tags),
		track.UpdatedAt,
	)

	if err != nil {
		// TODO: Check unique constraint violation
		r.logger.ErrorContext(ctx, "Error updating audio track", "error", err, "trackID", track.ID)
		return fmt.Errorf("updating audio track: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	r.logger.InfoContext(ctx, "Audio track updated successfully", "trackID", track.ID)
	return nil
}

func (r *AudioTrackRepository) Delete(ctx context.Context, id domain.TrackID) error {
	query := `DELETE FROM audio_tracks WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting audio track", "error", err, "trackID", id)
		return fmt.Errorf("deleting audio track: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound // Not found
	}
	r.logger.InfoContext(ctx, "Audio track deleted successfully", "trackID", id)
	return nil
}

func (r *AudioTrackRepository) Exists(ctx context.Context, id domain.TrackID) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM audio_tracks WHERE id = $1)`
    err := r.db.QueryRow(ctx, query, id).Scan(&exists)
    if err != nil {
        r.logger.ErrorContext(ctx, "Error checking audio track existence", "error", err, "trackID", id)
        return false, fmt.Errorf("checking track existence: %w", err)
    }
    return exists, nil
}


// --- Helper Methods ---

// RowScanner defines an interface satisfied by both pgx.Row and pgx.Rows.
type RowScanner interface {
	Scan(dest ...any) error
}

// scanTrack scans a single row into a domain.AudioTrack.
func (r *AudioTrackRepository) scanTrack(ctx context.Context, row RowScanner) (*domain.AudioTrack, error) {
	var track domain.AudioTrack
	var langCode string
	var durationMs int64 // Scan into int64 first
	var tags pq.StringArray // Scan TEXT[] into pq.StringArray

	err := row.Scan(
		&track.ID,
		&track.Title,
		&track.Description,
		&langCode, // Scan language code
		&track.Level,
		&durationMs, // Scan duration_ms
		&track.MinioBucket,
		&track.MinioObjectKey,
		&track.CoverImageURL,
		&track.UploaderID, // Scans correctly into *domain.UserID (*uuid.UUID)
		&track.IsPublic,
		&tags, // Scan into pq.StringArray
		&track.CreatedAt,
		&track.UpdatedAt,
	)
	if err != nil {
		return nil, err // Let caller handle pgx.ErrNoRows or other DB errors
	}

	// Convert scanned values to domain types
	langVO, langErr := domain.NewLanguage(langCode, "") // Name is not stored, can be looked up later if needed
	if langErr != nil {
		r.logger.WarnContext(ctx, "Invalid language code found in database", "error", langErr, "langCode", langCode, "trackID", track.ID)
		// Decide how to handle - return error or default language? Using default for now.
		// langVO = domain.Language{} // Or a specific unknown language
		return nil, fmt.Errorf("invalid language code %s in DB for track %s: %w", langCode, track.ID, langErr)
	}
	track.Language = langVO
	track.Duration = time.Duration(durationMs) * time.Millisecond // Convert ms to time.Duration
	track.Tags = tags // Assign []string from pq.StringArray

	return &track, nil
}

// Compile-time check
var _ port.AudioTrackRepository = (*AudioTrackRepository)(nil)