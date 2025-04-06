// internal/usecase/audio_content_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config" // Adjust import path (for presign expiry)
	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // GetUserID
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"                      // Import pagination
)

// AudioContentUseCase handles business logic related to audio tracks and collections.
type AudioContentUseCase struct {
	trackRepo      port.AudioTrackRepository
	collectionRepo port.AudioCollectionRepository
	storageService port.FileStorageService
	txManager      port.TransactionManager // ADDED: Transaction Manager
	presignExpiry  time.Duration
	logger         *slog.Logger
}

// NewAudioContentUseCase creates a new AudioContentUseCase.
func NewAudioContentUseCase(
	minioCfg config.MinioConfig,
	tr port.AudioTrackRepository,
	cr port.AudioCollectionRepository,
	ss port.FileStorageService,
	tm port.TransactionManager, // ADDED: Transaction Manager dependency
	log *slog.Logger,
) *AudioContentUseCase {
	if tm == nil {
		log.Warn("AudioContentUseCase created without TransactionManager implementation. Transactional operations will fail.")
	}
	return &AudioContentUseCase{
		trackRepo:      tr,
		collectionRepo: cr,
		storageService: ss,
		txManager:      tm, // Store it
		presignExpiry:  minioCfg.PresignExpiry,
		logger:         log.With("usecase", "AudioContentUseCase"),
	}
}

// --- Track Use Cases ---

// GetAudioTrackDetails retrieves details for a single audio track, including a presigned URL.
func (uc *AudioContentUseCase) GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error) {
	track, err := uc.trackRepo.FindByID(ctx, trackID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Audio track not found", "trackID", trackID)
		} else {
			uc.logger.ErrorContext(ctx, "Failed to get audio track from repository", "error", err, "trackID", trackID)
		}
		return nil, "", err
	}

	// Optional: Authorization Check (Example)
	/*
	if !track.IsPublic {
		userID, ok := middleware.GetUserIDFromContext(ctx)
		if !ok { return nil, "", domain.ErrUnauthenticated }
		if track.UploaderID == nil || *track.UploaderID != userID {
			return nil, "", domain.ErrPermissionDenied
		}
	}
	*/

	playURL, err := uc.storageService.GetPresignedGetURL(ctx, track.MinioBucket, track.MinioObjectKey, uc.presignExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate presigned URL for track", "error", err, "trackID", trackID)
		return nil, "", fmt.Errorf("could not retrieve playback URL: %w", err)
	}

	uc.logger.InfoContext(ctx, "Successfully retrieved audio track details", "trackID", trackID)
	return track, playURL, nil
}

// ListTracks retrieves a paginated list of tracks based on filtering and sorting criteria.
func (uc *AudioContentUseCase) ListTracks(ctx context.Context, params port.ListTracksParams, limit, offset int) ([]*domain.AudioTrack, int, pagination.Page, error) {
	pageParams := pagination.NewPageFromOffset(limit, offset)
	tracks, total, err := uc.trackRepo.List(ctx, params, pageParams)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list audio tracks from repository", "error", err, "params", params, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve track list: %w", err)
	}
	uc.logger.InfoContext(ctx, "Successfully listed audio tracks", "count", len(tracks), "total", total, "params", params, "page", pageParams)
	return tracks, total, pageParams, nil
}

// --- Collection Use Cases ---

// CreateCollection creates a new audio collection, potentially adding initial tracks atomically.
func (uc *AudioContentUseCase) CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return nil, domain.ErrUnauthenticated }

	if uc.txManager == nil {
		uc.logger.ErrorContext(ctx, "TransactionManager is nil, cannot create collection atomically")
		return nil, fmt.Errorf("internal configuration error: transaction manager not available")
	}


	collection, err := domain.NewAudioCollection(title, description, userID, colType)
	if err != nil {
		uc.logger.WarnContext(ctx, "Failed to create collection domain object", "error", err, "title", title, "type", colType, "userID", userID)
		return nil, err
	}

	// Execute repository operations within a transaction
	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		// 1. Save collection metadata within the transaction
		if err := uc.collectionRepo.Create(txCtx, collection); err != nil {
			return fmt.Errorf("saving collection metadata: %w", err)
		}

		// 2. Add initial tracks if provided, also within the transaction
		if len(initialTrackIDs) > 0 {
			// Optional: Validate track IDs exist before attempting insertion
			exists, validateErr := uc.validateTrackIDsExist(txCtx, initialTrackIDs)
			if validateErr != nil {
				// Logged inside validateTrackIDsExist
				return fmt.Errorf("validating initial tracks: %w", validateErr) // Propagate validation error
			}
			if !exists {
				// This condition implies specific IDs were missing, error logged inside validate func
				// The specific error message should come from validateTrackIDsExist ideally
				return fmt.Errorf("%w: one or more initial track IDs do not exist", domain.ErrInvalidArgument)
			}

			// Call ManageTracks repo method within the transaction context
			if err := uc.collectionRepo.ManageTracks(txCtx, collection.ID, initialTrackIDs); err != nil {
				return fmt.Errorf("adding initial tracks: %w", err)
			}
			// Update in-memory object only on success within transaction scope
			collection.TrackIDs = initialTrackIDs
		}
		return nil // Commit transaction
	})

	if finalErr != nil {
		uc.logger.ErrorContext(ctx, "Transaction failed during collection creation", "error", finalErr, "collectionID", collection.ID, "userID", userID)
		return nil, fmt.Errorf("failed to create collection: %w", finalErr)
	}

	uc.logger.InfoContext(ctx, "Audio collection created", "collectionID", collection.ID, "userID", userID)
	return collection, nil
}

// GetCollectionDetails retrieves details for a single collection, including its ordered track list.
func (uc *AudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx)

	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) { uc.logger.WarnContext(ctx, "Audio collection not found", "collectionID", collectionID) } else { uc.logger.ErrorContext(ctx, "Failed to get audio collection from repository", "error", err, "collectionID", collectionID) }
		return nil, err
	}

	// Authorization Check: Only owner can view (example rule)
	if !userAuthenticated || collection.OwnerID != userID {
		uc.logger.WarnContext(ctx, "Permission denied for user on collection", "collectionID", collectionID, "ownerID", collection.OwnerID, "requestingUserID", userID)
		return nil, domain.ErrPermissionDenied
	}

	uc.logger.InfoContext(ctx, "Successfully retrieved collection details", "collectionID", collectionID, "trackCount", len(collection.TrackIDs))
	return collection, nil
}

// GetCollectionTracks retrieves the ordered list of tracks for a specific collection.
func (uc *AudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
	// NOTE: Assumes authorization check happened before (e.g., in GetCollectionDetails or middleware)
	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
    if err != nil { return nil, err } // Handles NotFound

    if len(collection.TrackIDs) == 0 { return []*domain.AudioTrack{}, nil }

    tracks, err := uc.trackRepo.ListByIDs(ctx, collection.TrackIDs)
    if err != nil {
        uc.logger.ErrorContext(ctx, "Failed to list tracks by IDs for collection", "error", err, "collectionID", collectionID)
        return nil, fmt.Errorf("failed to retrieve track details for collection: %w", err)
    }
    uc.logger.InfoContext(ctx, "Successfully retrieved tracks for collection", "collectionID", collectionID, "trackCount", len(tracks))
    return tracks, nil
}


// UpdateCollectionMetadata updates the title and description of a collection owned by the user.
func (uc *AudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	if title == "" { return fmt.Errorf("%w: collection title cannot be empty", domain.ErrInvalidArgument) }

	// Repo update includes owner check in WHERE clause
	tempCollection := &domain.AudioCollection{ ID: collectionID, OwnerID: userID, Title: title, Description: description }
	err := uc.collectionRepo.UpdateMetadata(ctx, tempCollection)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrPermissionDenied) {
			uc.logger.ErrorContext(ctx, "Failed to update collection metadata", "error", err, "collectionID", collectionID, "userID", userID)
		}
		return err
	}
	uc.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collectionID, "userID", userID)
	return nil
}

// UpdateCollectionTracks updates the list and order of tracks atomically using TransactionManager.
func (uc *AudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	if uc.txManager == nil {
		uc.logger.ErrorContext(ctx, "TransactionManager is nil, cannot update collection tracks atomically")
		return fmt.Errorf("internal configuration error: transaction manager not available")
	}


	// Wrap the logic in a transaction
	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		// 1. Verify collection exists and user owns it (within Tx)
		collection, err := uc.collectionRepo.FindByID(txCtx, collectionID) // Use txCtx
		if err != nil { return err } // Propagate error (NotFound or DB error)
		if collection.OwnerID != userID {
			uc.logger.WarnContext(txCtx, "Attempt to modify tracks of collection not owned by user", "collectionID", collectionID, "ownerID", collection.OwnerID, "userID", userID)
			return domain.ErrPermissionDenied
		}

		// 2. Validate that all provided track IDs exist (within Tx)
		if len(orderedTrackIDs) > 0 {
			exists, validateErr := uc.validateTrackIDsExist(txCtx, orderedTrackIDs) // Use txCtx
			if validateErr != nil { return fmt.Errorf("validating tracks: %w", validateErr) }
			if !exists { return fmt.Errorf("%w: one or more track IDs do not exist", domain.ErrInvalidArgument) }
		}

		// 3. Call repository method to manage tracks (passes txCtx implicitly via getQuerier)
		if err := uc.collectionRepo.ManageTracks(txCtx, collectionID, orderedTrackIDs); err != nil {
			return fmt.Errorf("updating collection tracks in repository: %w", err)
		}
		return nil // Commit transaction
	})

	if finalErr != nil {
		uc.logger.ErrorContext(ctx, "Transaction failed during collection track update", "error", finalErr, "collectionID", collectionID, "userID", userID)
		return fmt.Errorf("failed to update collection tracks: %w", finalErr)
	}

	uc.logger.InfoContext(ctx, "Collection tracks updated", "collectionID", collectionID, "userID", userID, "trackCount", len(orderedTrackIDs))
	return nil
}


// DeleteCollection deletes a collection owned by the user.
func (uc *AudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	// 1. Verify ownership first (read operation, likely no Tx needed)
	collection, err := uc.collectionRepo.FindByID(ctx, collectionID)
	if err != nil { return err } // Handles ErrNotFound
	if collection.OwnerID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete collection not owned by user", "collectionID", collectionID, "ownerID", collection.OwnerID, "userID", userID)
		return domain.ErrPermissionDenied
	}

	// 2. Delete from repository (simple delete, likely no Tx needed unless linked hooks exist)
	err = uc.collectionRepo.Delete(ctx, collectionID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) { uc.logger.ErrorContext(ctx, "Failed to delete collection from repository", "error", err, "collectionID", collectionID, "userID", userID) }
		return err
	}
	uc.logger.InfoContext(ctx, "Collection deleted", "collectionID", collectionID, "userID", userID)
	return nil
}


// Helper function to validate track IDs exist
func (uc *AudioContentUseCase) validateTrackIDsExist(ctx context.Context, trackIDs []domain.TrackID) (bool, error) {
	if len(trackIDs) == 0 {
		return true, nil // Nothing to validate
	}
	existingTracks, err := uc.trackRepo.ListByIDs(ctx, trackIDs)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to validate track IDs existence", "error", err)
		return false, fmt.Errorf("failed to verify tracks: %w", err)
	}
	if len(existingTracks) != len(trackIDs) {
		// Find which IDs are missing (optional logging detail)
		providedSet := make(map[domain.TrackID]struct{}, len(trackIDs)); for _, id := range trackIDs { providedSet[id] = struct{}{} }; for _, track := range existingTracks { delete(providedSet, track.ID) }; missingIDs := make([]string, 0, len(providedSet)); for id := range providedSet { missingIDs = append(missingIDs, id.String()) }
		uc.logger.WarnContext(ctx, "Attempt to use non-existent track IDs", "missingTrackIDs", missingIDs)
		return false, nil // Return false, no error (usecase handles returning ErrInvalidArgument)
	}
	return true, nil
}

// Compile-time check to ensure AudioContentUseCase satisfies the port.AudioContentUseCase interface
var _ port.AudioContentUseCase = (*AudioContentUseCase)(nil)
