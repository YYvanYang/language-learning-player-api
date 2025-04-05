// internal/adapter/repository/postgres/audiocollection_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
)

type AudioCollectionRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewAudioCollectionRepository(db *pgxpool.Pool, logger *slog.Logger) *AudioCollectionRepository {
	return &AudioCollectionRepository{
		db:     db,
		logger: logger.With("repository", "AudioCollectionRepository"),
	}
}

// --- Interface Implementation ---

func (r *AudioCollectionRepository) Create(ctx context.Context, collection *domain.AudioCollection) error {
	query := `
        INSERT INTO audio_collections (id, title, description, owner_id, type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.Exec(ctx, query,
		collection.ID,
		collection.Title,
		collection.Description,
		collection.OwnerID,
		collection.Type,
		collection.CreatedAt,
		collection.UpdatedAt,
	)
	if err != nil {
		// TODO: Check foreign key constraint violation for owner_id?
		r.logger.ErrorContext(ctx, "Error creating audio collection", "error", err, "collectionID", collection.ID, "ownerID", collection.OwnerID)
		return fmt.Errorf("creating audio collection: %w", err)
	}
	// Note: This only creates the collection metadata, not the tracks association.
	// ManageTracks should be called separately if tracks are provided initially.
	r.logger.InfoContext(ctx, "Audio collection created successfully", "collectionID", collection.ID, "title", collection.Title)
	return nil
}

func (r *AudioCollectionRepository) FindByID(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	query := `
        SELECT id, title, description, owner_id, type, created_at, updated_at
        FROM audio_collections
        WHERE id = $1
    `
	collection, err := r.scanCollection(ctx, r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.logger.ErrorContext(ctx, "Error finding collection by ID", "error", err, "collectionID", id)
		return nil, fmt.Errorf("finding collection by ID: %w", err)
	}
	// This doesn't load tracks, use FindWithTracks for that.
	collection.TrackIDs = make([]domain.TrackID, 0) // Ensure slice is initialized
	return collection, nil
}

// FindWithTracks retrieves collection metadata AND the ordered list of track IDs.
func (r *AudioCollectionRepository) FindWithTracks(ctx context.Context, id domain.CollectionID) (*domain.AudioCollection, error) {
	// 1. Get collection metadata
	collection, err := r.FindByID(ctx, id) // Reuse FindByID
	if err != nil {
		return nil, err // Handles ErrNotFound already
	}

	// 2. Get ordered track IDs
	queryTracks := `
        SELECT track_id
        FROM collection_tracks
        WHERE collection_id = $1
        ORDER BY position ASC
    `
	rows, err := r.db.Query(ctx, queryTracks, id)
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


func (r *AudioCollectionRepository) ListByOwner(ctx context.Context, ownerID domain.UserID, page port.Page) ([]*domain.AudioCollection, int, error) {
    var args []interface{}
	argID := 1

	baseQuery := `
        SELECT id, title, description, owner_id, type, created_at, updated_at
        FROM audio_collections
    `
	countQuery := `SELECT count(*) FROM audio_collections`
	whereClause := fmt.Sprintf(" WHERE owner_id = $%d", argID)
	args = append(args, ownerID)
	argID++

	// Count total
	var total int
	err := r.db.QueryRow(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error counting collections by owner", "error", err, "ownerID", ownerID)
		return nil, 0, fmt.Errorf("counting collections by owner: %w", err)
	}

	if total == 0 {
		return []*domain.AudioCollection{}, 0, nil
	}

	// Build final query with sorting (default by created_at) and pagination
	orderByClause := " ORDER BY created_at DESC"
	paginationClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, page.Limit, page.Offset)

	finalQuery := baseQuery + whereClause + orderByClause + paginationClause

	rows, err := r.db.Query(ctx, finalQuery, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error listing collections by owner", "error", err, "ownerID", ownerID)
		return nil, 0, fmt.Errorf("listing collections by owner: %w", err)
	}
	defer rows.Close()

	collections := make([]*domain.AudioCollection, 0, page.Limit)
	for rows.Next() {
		collection, err := r.scanCollection(ctx, rows)
		if err != nil {
			r.logger.ErrorContext(ctx, "Error scanning collection in ListByOwner", "error", err)
			continue // Or return error
		}
		collection.TrackIDs = make([]domain.TrackID, 0) // Initialize empty slice
		collections = append(collections, collection)
	}

	if err = rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "Error iterating collection rows in ListByOwner", "error", err)
		return nil, 0, fmt.Errorf("iterating collection rows: %w", err)
	}

	return collections, total, nil
}


func (r *AudioCollectionRepository) UpdateMetadata(ctx context.Context, collection *domain.AudioCollection) error {
    collection.UpdatedAt = time.Now() // Ensure updated_at is set
    query := `
        UPDATE audio_collections SET
            title = $2, description = $3, updated_at = $4
        WHERE id = $1 AND owner_id = $5 -- Include owner_id for optimistic concurrency/auth check
    `
    cmdTag, err := r.db.Exec(ctx, query,
        collection.ID,
        collection.Title,
        collection.Description,
        collection.UpdatedAt,
        collection.OwnerID, // Pass owner ID for the WHERE clause
    )
    if err != nil {
        r.logger.ErrorContext(ctx, "Error updating collection metadata", "error", err, "collectionID", collection.ID)
        return fmt.Errorf("updating collection metadata: %w", err)
    }
    if cmdTag.RowsAffected() == 0 {
        // Could be not found OR wrong owner
		// Check if collection exists first to distinguish?
		exists, _ := r.exists(ctx, collection.ID)
		if !exists {
			return domain.ErrNotFound
		}
		// If it exists, it means owner didn't match (permission issue potentially caught at DB level)
        return domain.ErrPermissionDenied // Or ErrNotFound if we don't want to reveal existence
    }
    r.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collection.ID)
    return nil
}

// ManageTracks replaces the entire set of tracks associated with a collection.
func (r *AudioCollectionRepository) ManageTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	// Use a transaction to ensure atomicity
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.ErrorContext(ctx, "Failed to begin transaction for ManageTracks", "error", err, "collectionID", collectionID)
		return fmt.Errorf("beginning transaction: %w", err)
	}
	// Defer rollback in case of error, commit will override if successful
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx) // Rollback on panic
			panic(p) // Re-panic after rollback
		} else if err != nil {
			r.logger.WarnContext(ctx, "Rolling back transaction for ManageTracks due to error", "error", err, "collectionID", collectionID)
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				r.logger.ErrorContext(ctx, "Error rolling back transaction", "rollbackError", rbErr, "originalError", err)
			}
		} else {
			err = tx.Commit(ctx) // Commit if no error occurred
			if err != nil {
				r.logger.ErrorContext(ctx, "Error committing transaction for ManageTracks", "error", err, "collectionID", collectionID)
				err = fmt.Errorf("committing transaction: %w", err) // Set error to be returned
			} else {
				r.logger.InfoContext(ctx, "Successfully managed tracks for collection", "collectionID", collectionID, "trackCount", len(orderedTrackIDs))
			}
		}
	}()


	// 1. Delete existing tracks for the collection within the transaction
	deleteQuery := `DELETE FROM collection_tracks WHERE collection_id = $1`
	if _, err = tx.Exec(ctx, deleteQuery, collectionID); err != nil {
		r.logger.ErrorContext(ctx, "Failed to delete old tracks in ManageTracks transaction", "error", err, "collectionID", collectionID)
		return fmt.Errorf("deleting old tracks: %w", err) // Error will trigger rollback
	}

	// 2. Insert new tracks with correct positions within the transaction
	if len(orderedTrackIDs) > 0 {
		insertQuery := `
            INSERT INTO collection_tracks (collection_id, track_id, position)
            VALUES ($1, $2, $3)
        `
		// Prepare statement for batch insert
		// Note: pgx has more efficient batching mechanisms (using CopyFrom),
		// but a prepared statement loop is simpler here for moderate numbers of tracks.
		stmt, errPrep := tx.Prepare(ctx, "insert_collection_track", insertQuery)
        if errPrep != nil {
            err = fmt.Errorf("preparing track insert statement: %w", errPrep)
            return err // Trigger rollback
        }

		for i, trackID := range orderedTrackIDs {
			if _, err = tx.Exec(ctx, stmt.Name, collectionID, trackID, i); err != nil {
				// TODO: Check for foreign key violation if trackID doesn't exist?
				r.logger.ErrorContext(ctx, "Failed to insert track in ManageTracks transaction", "error", err, "collectionID", collectionID, "trackID", trackID, "position", i)
				err = fmt.Errorf("inserting track %s at position %d: %w", trackID, i, err)
				return err // Trigger rollback
			}
		}
	}

	// If we reach here without error, the defer func will commit the transaction.
	return err // Return the potential commit error or nil
}


func (r *AudioCollectionRepository) Delete(ctx context.Context, id domain.CollectionID) error {
    // Note: Ownership check should happen in the Usecase layer before calling this.
	// The ON DELETE CASCADE constraints will handle removing collection_tracks entries.
	query := `DELETE FROM audio_collections WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting audio collection", "error", err, "collectionID", id)
		return fmt.Errorf("deleting audio collection: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	r.logger.InfoContext(ctx, "Audio collection deleted successfully", "collectionID", id)
	return nil
}

// --- Helper Methods ---

// exists checks if a collection with the given ID exists.
func (r *AudioCollectionRepository) exists(ctx context.Context, id domain.CollectionID) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM audio_collections WHERE id = $1)`
    err := r.db.QueryRow(ctx, query, id).Scan(&exists)
    if err != nil {
        // Log the error but don't necessarily return it if the goal is just the boolean check
        r.logger.ErrorContext(ctx, "Error checking collection existence", "error", err, "collectionID", id)
        return false, fmt.Errorf("checking collection existence: %w", err) // Return error for robustness
    }
    return exists, nil
}


// scanCollection scans a single row into a domain.AudioCollection.
func (r *AudioCollectionRepository) scanCollection(ctx context.Context, row RowScanner) (*domain.AudioCollection, error) {
	var collection domain.AudioCollection
	err := row.Scan(
		&collection.ID,
		&collection.Title,
		&collection.Description,
		&collection.OwnerID,
		&collection.Type,
		&collection.CreatedAt,
		&collection.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	// TrackIDs are not scanned here, they are loaded separately if needed.
	return &collection, nil
}


// Compile-time check
var _ port.AudioCollectionRepository = (*AudioCollectionRepository)(nil)