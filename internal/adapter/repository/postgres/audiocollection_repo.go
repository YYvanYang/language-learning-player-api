// internal/adapter/repository/postgres/audiocollection_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/jackc/pgx/v5/pgconn" // Needed only if checking specific pg error codes

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination" // Import pagination
)

type AudioCollectionRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	// Use Querier interface to work with both pool and transaction
	getQuerier func(ctx context.Context) Querier
}

func NewAudioCollectionRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioCollectionRepository {
	repo := &AudioCollectionRepository{
		db:     db,
		logger: logger.With("repository", "AudioCollectionRepository"),
	}
	// Initialize the getQuerier helper function
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db) // Use the helper defined in tx_manager.go or here
	}
	return repo
}

// --- Interface Implementation ---

func (r *AudioCollectionRepository) Create(ctx context.Context, collection *domain.AudioCollection) error {
	q := r.getQuerier(ctx) // Get appropriate querier (pool or tx)
	query := `
        INSERT INTO audio_collections (id, title, description, owner_id, type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := q.Exec(ctx, query,
		collection.ID, collection.Title, collection.Description, collection.OwnerID,
		collection.Type, collection.CreatedAt, collection.UpdatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error creating audio collection", "error", err, "collectionID", collection.ID, "ownerID", collection.OwnerID)
		// Consider mapping specific DB errors (like FK violation) to domain errors if needed
		return fmt.Errorf("creating audio collection: %w", err)
	}
	r.logger.InfoContext(ctx, "Audio collection created successfully", "collectionID", collection.ID, "title", collection.Title)
	return nil
}

func (r *AudioCollectionRepository) FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT id, title, description, owner_id, type, created_at, updated_at
        FROM audio_collections
        WHERE id = $1
    `
	collection, err := r.scanCollection(ctx, q.QueryRow(ctx, query, id)) // Pass QueryRow
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrNotFound }
		r.logger.ErrorContext(ctx, "Error finding collection by ID", "error", err, "collectionID", id)
		return nil, fmt.Errorf("finding collection by ID: %w", err)
	}
	collection.TrackIDs = make([]domain.TrackID, 0) // Ensure slice is initialized
	return collection, nil
}

func (r *AudioCollectionRepository) FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	q := r.getQuerier(ctx) // Use querier from context for potential transaction

	// 1. Get collection metadata
	collection, err := r.FindByID(ctx, id) // Reuse FindByID which uses getQuerier
	if err != nil {
		return nil, err // Handles ErrNotFound already
	}

	// 2. Get ordered track IDs using the same querier (pool or tx)
	queryTracks := `
        SELECT track_id
        FROM collection_tracks
        WHERE collection_id = $1
        ORDER BY position ASC
    `
	rows, err := q.Query(ctx, queryTracks, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error fetching track IDs for collection", "error", err, "collectionID", id)
		return nil, fmt.Errorf("fetching tracks for collection %s: %w", id, err)
	}
	defer rows.Close()

	trackIDs := make([]domain.TrackID, 0)
	for rows.Next() {
		var trackID domain.TrackID
		if err := rows.Scan(&trackID); err != nil {
			r.logger.ErrorContext(ctx, "Error scanning track ID for collection", "error", err, "collectionID", id)
			return nil, fmt.Errorf("scanning track ID for collection %s: %w", id, err)
		}
		trackIDs = append(trackIDs, trackID)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating track IDs for collection", "error", err, "collectionID", id)
		return nil, fmt.Errorf("iterating track IDs for collection %s: %w", id, err)
	}

	collection.TrackIDs = trackIDs
	return collection, nil
}


func (r *AudioCollectionRepository) ListByOwner(ctx context.Context, ownerID domain.UserID, page pagination.Page) ([]*domain.AudioCollection, int, error) {
	q := r.getQuerier(ctx)
	args := []interface{}{ownerID}
	argID := 2 // Start arg numbering after ownerID ($1)
	baseQuery := ` FROM audio_collections WHERE owner_id = $1 `
	countQuery := `SELECT count(*) ` + baseQuery
	selectQuery := `SELECT id, title, description, owner_id, type, created_at, updated_at ` + baseQuery

	var total int
	err := q.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting collections by owner", "error", err, "ownerID", ownerID)
		return nil, 0, fmt.Errorf("counting collections by owner: %w", err)
	}
	if total == 0 { return []*domain.AudioCollection{}, 0, nil }

	orderByClause := " ORDER BY created_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)
	finalQuery := selectQuery + orderByClause + paginationClause

	rows, err := q.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing collections by owner", "error", err, "ownerID", ownerID)
		return nil, 0, fmt.Errorf("listing collections by owner: %w", err)
	}
	defer rows.Close()
	collections := make([]*domain.AudioCollection, 0, page.Limit)
	for rows.Next() {
		collection, err := r.scanCollection(ctx, rows)
		if err != nil { r.logger.ErrorContext(ctx, "Error scanning collection in ListByOwner", "error", err); continue }
		collection.TrackIDs = make([]domain.TrackID, 0)
		collections = append(collections, collection)
	}
	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating collection rows in ListByOwner", "error", err)
		return nil, 0, fmt.Errorf("iterating collection rows: %w", err)
	}
	return collections, total, nil
}

// UpdateMetadata updates only title and description. Ownership check via WHERE clause.
func (r *AudioCollectionRepository) UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error {
	q := r.getQuerier(ctx)
	collection.UpdatedAt = time.Now()
	query := `
        UPDATE audio_collections SET
            title = $2, description = $3, updated_at = $4
        WHERE id = $1 AND owner_id = $5 -- Ensure owner matches
    `
	cmdTag, err := q.Exec(ctx, query,
		collection.ID, collection.Title, collection.Description, collection.UpdatedAt, collection.OwnerID,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error updating collection metadata", "error", err, "collectionID", collection.ID)
		return fmt.Errorf("updating collection metadata: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		exists, _ := r.exists(ctx, collection.ID) // Check existence
		if !exists { return domain.ErrNotFound }
		return domain.ErrPermissionDenied // Assume owner mismatch if exists but not updated
	}
	r.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collection.ID)
	return nil
}


// ManageTracks replaces the entire set of tracks associated with a collection.
// This method now expects to run within a transaction context provided by the Usecase.
func (r *AudioCollectionRepository) ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	q := r.getQuerier(ctx) // Gets Tx if called within TxManager.Execute, otherwise Pool

	// 1. Delete existing tracks for the collection
	deleteQuery := `DELETE FROM collection_tracks WHERE collection_id = $1`
	if _, err := q.Exec(ctx, deleteQuery, collectionID); err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete old tracks in ManageTracks", "error", err, "collectionID", collectionID)
		return fmt.Errorf("deleting old tracks: %w", err)
	}

	// 2. Insert new tracks with correct positions
	if len(orderedTrackIDs) > 0 {
		insertQuery := `
            INSERT INTO collection_tracks (collection_id, track_id, position)
            VALUES ($1, $2, $3)
        `
		// Use pgx's Batch for potentially better performance than looping Exec
		batch := &pgx.Batch{}
		for i, trackID := range orderedTrackIDs {
			batch.Queue(insertQuery, collectionID, trackID, i)
		}

		br := q.SendBatch(ctx, batch)
		defer br.Close() // Ensure batch results are closed

		// Check results for each queued insert
		for i := 0; i < len(orderedTrackIDs); i++ {
			_, err := br.Exec()
			if err != nil {
				r.logger.ErrorContext(ctx, "Failed to insert track in ManageTracks batch", "error", err, "collectionID", collectionID, "position", i)
				// Consider checking for FK violation on track_id here
				return fmt.Errorf("inserting track at position %d: %w", i, err)
			}
		}
	}

	// 3. Update the collection's updated_at timestamp
	updateTsQuery := `UPDATE audio_collections SET updated_at = $1 WHERE id = $2`
	if _, err := q.Exec(ctx, updateTsQuery, time.Now(), collectionID); err != nil {
		r.logger.ErrorContext(ctx, "Failed to update collection timestamp in ManageTracks", "error", err, "collectionID", collectionID)
		return fmt.Errorf("updating collection timestamp: %w", err)
	}

	r.logger.DebugContext(ctx, "ManageTracks repo operations completed (within transaction)", "collectionID", collectionID, "trackCount", len(orderedTrackIDs))
	return nil // Usecase layer handles commit/rollback
}


func (r *AudioCollectionRepository) Delete(ctx context.Context, id domain.CollectionID) error {
	q := r.getQuerier(ctx)
	// Ownership check is done in Usecase layer
	// ON DELETE CASCADE handles collection_tracks entries automatically
	query := `DELETE FROM audio_collections WHERE id = $1`
	cmdTag, err := q.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting audio collection", "error", err, "collectionID", id)
		return fmt.Errorf("deleting audio collection: %w", err)
	}
	if cmdTag.RowsAffected() == 0 { return domain.ErrNotFound }
	r.logger.InfoContext(ctx, "Audio collection deleted successfully", "collectionID", id)
	return nil
}

// --- Helper Methods ---

// exists checks if a collection with the given ID exists.
func (r *AudioCollectionRepository) exists(ctx context.Context, id domain.CollectionID) (bool, error) {
	q := r.getQuerier(ctx)
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM audio_collections WHERE id = $1)`
	err := q.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error checking collection existence", "error", err, "collectionID", id)
		return false, fmt.Errorf("checking collection existence: %w", err)
	}
	return exists, nil
}


// scanCollection scans a single row into a domain.AudioCollection.
func (r *AudioCollectionRepository) scanCollection(ctx context.Context, row RowScanner) (*domain.AudioCollection, error) {
	var collection domain.AudioCollection
	err := row.Scan(
		&collection.ID, &collection.Title, &collection.Description, &collection.OwnerID,
		&collection.Type, &collection.CreatedAt, &collection.UpdatedAt,
	)
	if err != nil { return nil, err }
	return &collection, nil
}

// Ensure implementation satisfies the interface
var _ port.AudioCollectionRepository = (*AudioCollectionRepository)(nil)
