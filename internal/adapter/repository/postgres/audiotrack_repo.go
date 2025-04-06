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
	"github.com/jackc/pgx/v5/pgconn"   // Import
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq" // Using lib/pq for array handling with pgx and error codes
	"github.com/google/uuid" // For uuid.NullUUID

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"                       // Import pagination
)

type AudioTrackRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// Use Querier interface to work with both pool and transaction
	getQuerier func(ctx context.Context) Querier
}

func NewAudioTrackRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioTrackRepository {
	repo := &AudioTrackRepository{
		db:     db,
		logger: logger.With("repository", "AudioTrackRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Use the helper defined in tx_manager.go or here
	}
	return repo
}

// --- Interface Implementation ---

func (r *AudioTrackRepository) Create(ctx context.Context, track *domain.AudioTrack) error {
	q := r.getQuerier(ctx)
	query := `
		INSERT INTO audio_tracks
			(id, title, description, language_code, level, duration_ms,
			 minio_bucket, minio_object_key, cover_image_url, uploader_id,
			 is_public, tags, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := q.Exec(ctx, query,
		track.ID,
		track.Title,
		track.Description,
		track.Language.Code(),      // Use code from Language VO
		track.Level,                // Use AudioLevel directly (string)
		track.Duration.Milliseconds(), // CORRECTED: Store as milliseconds BIGINT
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "audio_tracks_minio_object_key_key") {
				return fmt.Errorf("creating audio track: %w: object key '%s' already exists", domain.ErrConflict, track.MinioObjectKey)
			}
			r.logger.WarnContext(ctx, "Unique constraint violation on audio track creation", "constraint", pgErr.ConstraintName, "trackID", track.ID)
			return fmt.Errorf("creating audio track: %w: resource conflict on unique field", domain.ErrConflict)
		}
		if errors.As(err, &pgErr) && pgErr.Code == ForeignKeyViolation {
			return fmt.Errorf("creating audio track: %w: referenced resource not found", domain.ErrInvalidArgument)
		}
		r.logger.ErrorContext(ctx, "Error creating audio track", "error", err, "trackID", track.ID)
		return fmt.Errorf("creating audio track: %w", err)
	}
	r.logger.InfoContext(ctx, "Audio track created successfully", "trackID", track.ID, "title", track.Title)
	return nil
}

func (r *AudioTrackRepository) FindByID(ctx context.Context, id domain.TrackID) (*domain.AudioTrack, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = $1
    `
	track, err := r.scanTrack(ctx, q.QueryRow(ctx, query, id)) // Pass QueryRow directly
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
	q := r.getQuerier(ctx)
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
		ORDER BY array_position(text_to_uuid_array($1), id) -- Assuming a helper function or use Go sort
    `
	// NOTE: text_to_uuid_array is not a standard function, you might need to create it
	// or sort in Go code after fetching. Let's assume Go sort for now.
	querySimple := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = ANY($1)
    `
	rows, err := q.Query(ctx, querySimple, uuidStrs) // Use simpler query
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks by IDs", "error", err)
		return nil, fmt.Errorf("listing audio tracks by IDs: %w", err)
	}
	defer rows.Close()
	trackMap := make(map[domain.TrackID]*domain.AudioTrack, len(ids))
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning track in ListByIDs", "error", err)
			continue
		}
		trackMap[track.ID] = track
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track rows in ListByIDs", "error", err)
		return nil, fmt.Errorf("iterating track rows: %w", err)
	}
	// Re-order in Go code
	orderedTracks := make([]*domain.AudioTrack, 0, len(ids))
	for _, id := range ids {
		if t, ok := trackMap[id]; ok {
			orderedTracks = append(orderedTracks, t)
		}
	}
	return orderedTracks, nil
}

func (r *AudioTrackRepository) List(ctx context.Context, params port.ListTracksParams, page pagination.Page) ([]*domain.AudioTrack, int, error) {
	q := r.getQuerier(ctx)
	var args []interface{}
	argID := 1
	baseQuery := ` FROM audio_tracks `
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, title, description, language_code, level, duration_ms, minio_bucket, minio_object_key, cover_image_url, uploader_id, is_public, tags, created_at, updated_at ` + baseQuery
	whereClause := " WHERE 1=1"

	if params.Query != nil && *params.Query != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argID, argID)
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

	var total int
	err := q.QueryRow(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting audio tracks", "error", err, "filters", params)
		return nil, 0, fmt.Errorf("counting audio tracks: %w", err)
	}
	if total == 0 { return []*domain.AudioTrack{}, 0, nil }

	orderByClause := " ORDER BY created_at DESC" // Default sort
	if params.SortBy != "" {
		allowedSorts := map[string]string{"createdAt": "created_at", "title": "title", "duration": "duration_ms", "level": "level"}
		dbColumn, ok := allowedSorts[params.SortBy]
		if ok {
			direction := " ASC"; if strings.ToLower(params.SortDirection) == "desc" { direction = " DESC" }
			orderByClause = fmt.Sprintf(" ORDER BY %s%s", dbColumn, direction)
		} else { r.logger.WarnContext(ctx, "Invalid sort field requested", "sortBy", params.SortBy) }
	}
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)
	finalQuery := selectQuery + whereClause + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks", "error", err, "filters", params, "page", page)
		return nil, 0, fmt.Errorf("listing audio tracks: %w", err)
	}
	defer rows.Close()
	tracks := make([]*domain.AudioTrack, 0, page.Limit)
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows)
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning track in List", "error", err); continue }
		tracks = append(tracks, track)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track rows in List", "error", err)
		return nil, 0, fmt.Errorf("iterating track rows: %w", err)
	}
	return tracks, total, nil
}


func (r *AudioTrackRepository) Update(ctx context.Context, track *domain.AudioTrack) error {
	q := r.getQuerier(ctx)
	track.UpdatedAt = time.Now()
	query := `
		UPDATE audio_tracks SET
			title = $2, description = $3, language_code = $4, level = $5, duration_ms = $6,
			minio_bucket = $7, minio_object_key = $8, cover_image_url = $9, uploader_id = $10,
			is_public = $11, tags = $12, updated_at = $13
		WHERE id = $1
	`
	cmdTag, err := q.Exec(ctx, query,
		track.ID, track.Title, track.Description,
		track.Language.Code(),      // CORRECTED: Use VO code
		track.Level,                // CORRECTED: Use domain type directly
		track.Duration.Milliseconds(), // CORRECTED: Save milliseconds
		track.MinioBucket, track.MinioObjectKey, track.CoverImageURL, track.UploaderID,
		track.IsPublic, pq.Array(track.Tags), track.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "audio_tracks_minio_object_key_key") {
				return fmt.Errorf("updating audio track: %w: object key '%s' already exists", domain.ErrConflict, track.MinioObjectKey)
			}
			r.logger.WarnContext(ctx, "Unique constraint violation on audio track update", "constraint", pgErr.ConstraintName, "trackID", track.ID)
			return fmt.Errorf("updating audio track: %w: resource conflict on unique field", domain.ErrConflict)
		}
		r.logger.ErrorContext(ctx, "Error updating audio track", "error", err, "trackID", track.ID)
		return fmt.Errorf("updating audio track: %w", err)
	}
	if cmdTag.RowsAffected() == 0 { return domain.ErrNotFound }
	r.logger.InfoContext(ctx, "Audio track updated successfully", "trackID", track.ID)
	return nil
}

func (r *AudioTrackRepository) Delete(ctx context.Context, id domain.TrackID) error {
	q := r.getQuerier(ctx)
	query := `DELETE FROM audio_tracks WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting audio track", "error", err, "trackID", id)
		return fmt.Errorf("deleting audio track: %w", err)
	}
	if cmdTag.RowsAffected() == 0 { return domain.ErrNotFound }
	r.logger.InfoContext(ctx, "Audio track deleted successfully", "trackID", id)
	return nil
}

func (r *AudioTrackRepository) Exists(ctx context.Context, id domain.TrackID) (bool, error) {
	q := r.getQuerier(ctx)
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM audio_tracks WHERE id = $1)`
	err := q.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error checking audio track existence", "error", err, "trackID", id)
		return false, fmt.Errorf("checking track existence: %w", err)
	}
	return exists, nil
}

// scanTrack scans a single row into a domain.AudioTrack.
func (r *AudioTrackRepository) scanTrack(ctx context.Context, row RowScanner) (*domain.AudioTrack, error) {
	var track domain.AudioTrack
	var langCode string
	var levelStr string          // Scan level into string
	var durationMs int64         // CORRECTED: Scan into int64 for milliseconds
	var tags pq.StringArray      // Scan TEXT[] into pq.StringArray
	var uploaderID uuid.NullUUID // Use uuid.NullUUID for nullable foreign key

	err := row.Scan(
		&track.ID, &track.Title, &track.Description,
		&langCode,       // Scan language code
		&levelStr,       // Scan level string
		&durationMs,     // CORRECTED: Scan duration_ms
		&track.MinioBucket, &track.MinioObjectKey, &track.CoverImageURL,
		&uploaderID,     // Scan into NullUUID
		&track.IsPublic, &tags, &track.CreatedAt, &track.UpdatedAt,
	)
	if err != nil { return nil, err }

	// Convert scanned values to domain types
	langVO, langErr := domain.NewLanguage(langCode, "") // Name is not stored
	if langErr != nil {
		r.logger.ErrorContext(ctx, "Invalid language code found in database", "error", langErr, "langCode", langCode, "trackID", track.ID)
		return nil, fmt.Errorf("invalid language code %s in DB for track %s: %w", langCode, track.ID, langErr)
	}
	track.Language = langVO // CORRECTED: Assign to Language field
	track.Level = domain.AudioLevel(levelStr) // CORRECTED: Assign to Level field
	if !track.Level.IsValid() { // Validate scanned level
         r.logger.WarnContext(ctx, "Invalid audio level found in database", "level", levelStr, "trackID", track.ID)
         track.Level = domain.LevelUnknown // Assign default if invalid?
    }
	track.Duration = time.Duration(durationMs) * time.Millisecond // CORRECTED: Convert ms to time.Duration
	track.Tags = tags // Assign []string from pq.StringArray

	if uploaderID.Valid {
		uid := domain.UserID(uploaderID.UUID)
		track.UploaderID = &uid
	} else { track.UploaderID = nil }

	return &track, nil
}

var _ port.AudioTrackRepository = (*AudioTrackRepository)(nil)
