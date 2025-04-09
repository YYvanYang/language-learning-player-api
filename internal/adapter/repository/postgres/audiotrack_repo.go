// internal/adapter/repository/postgres/audiotrack_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq" // Using lib/pq for array handling with pgx and error codes

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
)

type AudioTrackRepository struct {
	db         *pgxpool.Pool
	logger     *slog.Logger
	getQuerier func(ctx context.Context) Querier
}

func NewAudioTrackRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioTrackRepository {
	repo := &AudioTrackRepository{
		db:     db,
		logger: logger.With("repository", "AudioTrackRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db)
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
		track.Language.Code(),
		track.Level,
		track.Duration.Milliseconds(), // Point 1: Convert domain Duration to int64 ms
		track.MinioBucket,
		track.MinioObjectKey,
		track.CoverImageURL,
		track.UploaderID,
		track.IsPublic,
		pq.Array(track.Tags),
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
	track, err := r.scanTrack(ctx, q.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
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
	uuidStrs := make([]string, len(ids))
	for i, id := range ids {
		uuidStrs[i] = id.String()
	}
	querySimple := `
        SELECT id, title, description, language_code, level, duration_ms,
               minio_bucket, minio_object_key, cover_image_url, uploader_id,
               is_public, tags, created_at, updated_at
        FROM audio_tracks
        WHERE id = ANY($1)
    `
	rows, err := q.Query(ctx, querySimple, uuidStrs)
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
	orderedTracks := make([]*domain.AudioTrack, 0, len(ids))
	for _, id := range ids {
		if t, ok := trackMap[id]; ok {
			orderedTracks = append(orderedTracks, t)
		}
	}
	return orderedTracks, nil
}

// Point 5: Updated to use filters port.ListTracksFilters
func (r *AudioTrackRepository) List(ctx context.Context, filters port.ListTracksFilters, page pagination.Page) ([]*domain.AudioTrack, int, error) {
	q := r.getQuerier(ctx)
	var args []interface{}
	argID := 1
	baseQuery := ` FROM audio_tracks `
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, title, description, language_code, level, duration_ms, minio_bucket, minio_object_key, cover_image_url, uploader_id, is_public, tags, created_at, updated_at ` + baseQuery
	whereClause := " WHERE 1=1"

	// Apply filters from ListTracksFilters
	if filters.Query != nil && *filters.Query != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argID, argID)
		args = append(args, "%"+*filters.Query+"%")
		argID++
	}
	if filters.LanguageCode != nil && *filters.LanguageCode != "" {
		whereClause += fmt.Sprintf(" AND language_code = $%d", argID)
		args = append(args, *filters.LanguageCode)
		argID++
	}
	if filters.Level != nil && *filters.Level != "" {
		whereClause += fmt.Sprintf(" AND level = $%d", argID)
		args = append(args, *filters.Level)
		argID++
	}
	if filters.IsPublic != nil {
		whereClause += fmt.Sprintf(" AND is_public = $%d", argID)
		args = append(args, *filters.IsPublic)
		argID++
	}
	if filters.UploaderID != nil {
		whereClause += fmt.Sprintf(" AND uploader_id = $%d", argID)
		args = append(args, *filters.UploaderID)
		argID++
	}
	if len(filters.Tags) > 0 {
		whereClause += fmt.Sprintf(" AND tags @> $%d", argID)
		args = append(args, pq.Array(filters.Tags))
		argID++
	}

	var total int
	err := q.QueryRow(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting audio tracks", "error", err, "filters", filters)
		return nil, 0, fmt.Errorf("counting audio tracks: %w", err)
	}
	if total == 0 {
		return []*domain.AudioTrack{}, 0, nil
	}

	orderByClause := " ORDER BY created_at DESC"
	if filters.SortBy != "" {
		// Point 1 & 5: Use duration_ms for sorting if specified
		allowedSorts := map[string]string{"createdAt": "created_at", "title": "title", "durationMs": "duration_ms", "level": "level"}
		dbColumn, ok := allowedSorts[filters.SortBy]
		if ok {
			direction := " ASC"
			if strings.ToLower(filters.SortDirection) == "desc" {
				direction = " DESC"
			}
			orderByClause = fmt.Sprintf(" ORDER BY %s%s", dbColumn, direction)
		} else {
			r.logger.WarnContext(ctx, "Invalid sort field requested", "sortBy", filters.SortBy)
		}
	}
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)
	finalQuery := selectQuery + whereClause + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing audio tracks", "error", err, "filters", filters, "page", page)
		return nil, 0, fmt.Errorf("listing audio tracks: %w", err)
	}
	defer rows.Close()
	tracks := make([]*domain.AudioTrack, 0, page.Limit)
	for rows.Next() {
		track, err := r.scanTrack(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning track in List", "error", err)
			continue
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
		track.Language.Code(),
		track.Level,
		track.Duration.Milliseconds(), // Point 1: Convert domain Duration to int64 ms
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
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
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
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
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

// Point 1: Updated scanTrack
func (r *AudioTrackRepository) scanTrack(ctx context.Context, row RowScanner) (*domain.AudioTrack, error) {
	var track domain.AudioTrack
	var langCode string
	var levelStr string
	var durationMs int64 // Scan into int64
	var tags pq.StringArray
	var uploaderID uuid.NullUUID

	err := row.Scan(
		&track.ID, &track.Title, &track.Description,
		&langCode,
		&levelStr,
		&durationMs, // Scan duration_ms column
		&track.MinioBucket, &track.MinioObjectKey, &track.CoverImageURL,
		&uploaderID,
		&track.IsPublic, &tags, &track.CreatedAt, &track.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	langVO, langErr := domain.NewLanguage(langCode, "")
	if langErr != nil {
		r.logger.ErrorContext(ctx, "Invalid language code found in database", "error", langErr, "langCode", langCode, "trackID", track.ID)
		return nil, fmt.Errorf("invalid language code %s in DB for track %s: %w", langCode, track.ID, langErr)
	}
	track.Language = langVO
	track.Level = domain.AudioLevel(levelStr)
	if !track.Level.IsValid() {
		r.logger.WarnContext(ctx, "Invalid audio level found in database", "level", levelStr, "trackID", track.ID)
		track.Level = domain.LevelUnknown
	}
	// Convert scanned milliseconds back to time.Duration
	track.Duration = time.Duration(durationMs) * time.Millisecond
	track.Tags = tags

	if uploaderID.Valid {
		uid := domain.UserID(uploaderID.UUID)
		track.UploaderID = &uid
	} else {
		track.UploaderID = nil
	}

	return &track, nil
}

var _ port.AudioTrackRepository = (*AudioTrackRepository)(nil)
