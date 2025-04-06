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
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // 添加middleware包
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"                      // Import pagination
)

// AudioContentUseCase handles business logic related to audio tracks and collections.
type AudioContentUseCase struct {
	trackRepo      port.AudioTrackRepository
	collectionRepo port.AudioCollectionRepository
	storageService port.FileStorageService
	// progressRepo   port.PlaybackProgressRepository // Optional: Inject if needed for GetDetails
	// bookmarkRepo   port.BookmarkRepository     // Optional: Inject if needed for GetDetails
	presignExpiry  time.Duration
	logger         *slog.Logger
}

// NewAudioContentUseCase creates a new AudioContentUseCase.
func NewAudioContentUseCase(
	minioCfg config.MinioConfig, // Pass Minio config for default expiry
	tr port.AudioTrackRepository,
	cr port.AudioCollectionRepository,
	ss port.FileStorageService,
	log *slog.Logger,
	// Optional dependencies:
	// pr port.PlaybackProgressRepository,
	// br port.BookmarkRepository,
) *AudioContentUseCase {
	return &AudioContentUseCase{
		trackRepo:      tr,
		collectionRepo: cr,
		storageService: ss,
		// progressRepo:   pr,
		// bookmarkRepo:   br,
		presignExpiry:  minioCfg.PresignExpiry, // Get default expiry from config
		logger:         log.With("usecase", "AudioContentUseCase"),
	}
}

// --- Track Use Cases ---

// GetAudioTrackDetails retrieves details for a single audio track, including a presigned URL.
// Assumes public access or authentication/authorization handled before calling.
func (uc *AudioContentUseCase) GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error) {
	// 1. Get track metadata from repository
	track, err := uc.trackRepo.FindByID(ctx, trackID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Audio track not found", "trackID", trackID)
		} else {
			uc.logger.ErrorContext(ctx, "Failed to get audio track from repository", "error", err, "trackID", trackID)
		}
		return nil, "", err // Return original error (includes ErrNotFound)
	}

	// 2. Check authorization (Example - adapt based on actual rules)
	// Typically, authorization middleware handles this before the handler calls the usecase.
	// If specific logic is needed here (e.g., checking ownership for private tracks):
	/*
	if !track.IsPublic {
		userID, ok := middleware.GetUserIDFromContext(ctx) // Assuming middleware puts UserID in context
		if !ok {
			// This case means an unauthenticated user is trying to access a non-public track
			uc.logger.WarnContext(ctx,"Unauthenticated access attempt on private track", "trackID", trackID)
			return nil, "", domain.ErrUnauthenticated // Or ErrPermissionDenied
		}
		// Add check if userID is owner or has specific rights, e.g.:
		// if track.UploaderID == nil || *track.UploaderID != userID {
		//    // Check if user has access via a purchased course or subscription, etc.
		//    hasAccess := uc.checkUserAccessRights(ctx, userID, trackID)
		//    if !hasAccess {
		// 		uc.logger.WarnContext(ctx,"Permission denied for user on private track", "trackID", trackID, "userID", userID)
		// 		return nil, "", domain.ErrPermissionDenied
		//    }
		// }
	}
	*/

	// 3. Generate presigned URL for playback
	playURL, err := uc.storageService.GetPresignedGetURL(ctx, track.MinioBucket, track.MinioObjectKey, uc.presignExpiry)
	if err != nil {
		// Log the error, but decide if it's critical. Maybe return track details without URL?
		uc.logger.ErrorContext(ctx, "Failed to generate presigned URL for track", "error", err, "trackID", trackID, "bucket", track.MinioBucket, "key", track.MinioObjectKey)
		// For a player app, the play URL is essential, so let's return the error.
		return nil, "", fmt.Errorf("could not retrieve playback URL: %w", err)
	}

	// 4. (Optional) Fetch related data like user progress or bookmarks
	// Requires injecting progressRepo/bookmarkRepo and getting userID from context.
	/*
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx)
	var userProgress *domain.PlaybackProgress
	var userBookmarks []*domain.Bookmark
	if userAuthenticated {
		userProgress, _ = uc.progressRepo.Find(ctx, userID, trackID) // Ignore error if not found
		userBookmarks, _ = uc.bookmarkRepo.ListByUserAndTrack(ctx, userID, trackID) // Ignore error
	}
	// How to return this extra data? Add fields to a DTO or return multiple values?
	// Returning only track and URL for now, handler can call other usecases if needed.
	*/

	uc.logger.InfoContext(ctx, "Successfully retrieved audio track details", "trackID", trackID)
	return track, playURL, nil
}


// ListTracks retrieves a paginated list of tracks based on filtering and sorting criteria.
// It receives raw limit/offset from the handler, creates a pagination.Page object, and returns it along with results.
func (uc *AudioContentUseCase) ListTracks(ctx context.Context, params port.ListTracksParams, limit, offset int) ([]*domain.AudioTrack, int, pagination.Page, error) {
	// Create Page object, applying defaults and constraints
	pageParams := pagination.NewPageFromOffset(limit, offset)

	// Ensure only public tracks are listed if no specific filter is set and user context doesn't grant special access
	// This logic might be better placed in the handler or middleware depending on complexity.
	// If the request itself can specify `isPublic=false`, ensure authorization.
	// Assuming basic listing shows public by default or respects filter:
	if params.IsPublic == nil {
		// Maybe default to only public? Depends on requirements.
		// isPublicTrue := true
		// params.IsPublic = &isPublicTrue
	}


	tracks, total, err := uc.trackRepo.List(ctx, params, pageParams) // Pass validated Page object
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list audio tracks from repository", "error", err, "params", params, "page", pageParams)
		// Return the validated pageParams even on error
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve track list: %w", err)
	}

	uc.logger.InfoContext(ctx, "Successfully listed audio tracks", "count", len(tracks), "total", total, "params", params, "page", pageParams)
	// Return tracks, total, and the validated pageParams used for the query
	return tracks, total, pageParams, nil
}


// --- Collection Use Cases ---

// CreateCollection creates a new audio collection for the currently authenticated user.
func (uc *AudioContentUseCase) CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx) // Assumes middleware provides UserID
	if !ok {
		return nil, domain.ErrUnauthenticated // Must be logged in to create collection
	}

	// 1. Create domain object
	collection, err := domain.NewAudioCollection(title, description, userID, colType)
	if err != nil {
		// Should only be invalid argument errors from domain constructor
		uc.logger.WarnContext(ctx, "Failed to create collection domain object", "error", err, "title", title, "type", colType, "userID", userID)
		return nil, err
	}

	// 2. Save metadata to repository
	if err := uc.collectionRepo.Create(ctx, collection); err != nil {
		uc.logger.ErrorContext(ctx, "Failed to save new collection metadata", "error", err, "collectionID", collection.ID, "userID", userID)
		return nil, fmt.Errorf("failed to create collection: %w", err) // Internal error
	}

	// 3. (Optional) Add initial tracks if provided
	if len(initialTrackIDs) > 0 {
		// TODO: Validate that all initialTrackIDs actually exist using trackRepo.ListByIDs or Exists?
		// For simplicity, assume valid IDs for now, DB foreign key will catch non-existent ones.
		if err := uc.collectionRepo.ManageTracks(ctx, collection.ID, initialTrackIDs); err != nil {
			uc.logger.ErrorContext(ctx, "Failed to add initial tracks to new collection", "error", err, "collectionID", collection.ID, "trackIDs", initialTrackIDs)
			// Should we delete the collection metadata if tracks fail? Or just log?
			// For now, log and return the collection metadata anyway, but with an error indicating tracks failed.
			return collection, fmt.Errorf("collection created, but failed to add initial tracks: %w", err)
		}
		collection.TrackIDs = initialTrackIDs // Update domain object in memory
	}

	uc.logger.InfoContext(ctx, "Audio collection created", "collectionID", collection.ID, "userID", userID)
	return collection, nil
}

// GetCollectionDetails retrieves details for a single collection, including its ordered track list.
// Checks ownership for access.
func (uc *AudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx)

	// 1. Get collection with tracks
	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Audio collection not found", "collectionID", collectionID)
		} else {
			uc.logger.ErrorContext(ctx, "Failed to get audio collection from repository", "error", err, "collectionID", collectionID)
		}
		return nil, err // Return original error
	}

	// 2. Authorization Check: Only owner can view (example rule)
	if !userAuthenticated || collection.OwnerID != userID {
		// Add more complex checks if needed (e.g., public collections, shared collections)
		uc.logger.WarnContext(ctx, "Permission denied for user on collection", "collectionID", collectionID, "ownerID", collection.OwnerID, "requestingUserID", userID)
		return nil, domain.ErrPermissionDenied // Or ErrNotFound if hiding existence is preferred
	}

	uc.logger.InfoContext(ctx, "Successfully retrieved collection details", "collectionID", collectionID, "trackCount", len(collection.TrackIDs))
	return collection, nil
}


// UpdateCollectionMetadata updates the title and description of a collection owned by the user.
func (uc *AudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	// 1. Get existing collection to verify ownership (or let repo handle it)
	// repo.UpdateMetadata includes owner_id in WHERE clause, so it implicitly checks ownership.

	// 2. Create a temporary domain object with updated fields
	// We need OwnerID for the repo check, even though we aren't changing it.
	// This feels a bit awkward. Maybe UpdateMetadata should only take fields to update?
	// Alternative: Repo method `UpdateMetadata(ctx, id, ownerId, title, desc)`
	tempCollection := &domain.AudioCollection{
		ID:          collectionID,
		OwnerID:     userID, // Crucial for ownership check in repo query
		Title:       title,
		Description: description,
		// Other fields like Type, TrackIDs, CreatedAt are not needed for this specific update
	}

	// Validate title (e.g., not empty)
	if title == "" {
		return fmt.Errorf("%w: collection title cannot be empty", domain.ErrInvalidArgument)
	}

	// 3. Call repository update
	err := uc.collectionRepo.UpdateMetadata(ctx, tempCollection)
	if err != nil {
		// Repo handles ErrNotFound and potential ErrPermissionDenied from WHERE clause check
		if !errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrPermissionDenied) {
			uc.logger.ErrorContext(ctx, "Failed to update collection metadata", "error", err, "collectionID", collectionID, "userID", userID)
		}
		// Return repo error directly
		return err
	}

	uc.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collectionID, "userID", userID)
	return nil
}

// GetCollectionTracks retrieves the ordered list of tracks for a specific collection.
// Assumes authorization check for the collection happened before calling (e.g., in GetCollectionDetails).
func (uc *AudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
    // 1. Get the collection first to get the track IDs in order
    // We could optimize this by having a repo method that directly returns ordered tracks,
    // but using FindWithTracks keeps logic simpler for now.
    collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
    if err != nil {
        // Log errors already handled in FindWithTracks call if needed
        return nil, err // Propagate NotFound etc.
    }

    // 2. Fetch the actual track details based on the IDs
    if len(collection.TrackIDs) == 0 {
        return []*domain.AudioTrack{}, nil // Return empty slice if no tracks
    }

    tracks, err := uc.trackRepo.ListByIDs(ctx, collection.TrackIDs)
    if err != nil {
        uc.logger.ErrorContext(ctx, "Failed to list tracks by IDs for collection", "error", err, "collectionID", collectionID)
        return nil, fmt.Errorf("failed to retrieve track details for collection: %w", err)
    }

	// ListByIDs should ideally return tracks in the requested order.
	// If not guaranteed, re-order here based on collection.TrackIDs.

    uc.logger.InfoContext(ctx, "Successfully retrieved tracks for collection", "collectionID", collectionID, "trackCount", len(tracks))
    return tracks, nil
}

// UpdateCollectionTracks updates the list and order of tracks in a collection owned by the user.
func (uc *AudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	// 1. Verify collection exists and user owns it
	collection, err := uc.collectionRepo.FindByID(ctx, collectionID) // Just need metadata for owner check
	if err != nil {
		return err // Handles ErrNotFound
	}
	if collection.OwnerID != userID {
		uc.logger.WarnContext(ctx, "Attempt to modify tracks of collection not owned by user", "collectionID", collectionID, "ownerID", collection.OwnerID, "userID", userID)
		return domain.ErrPermissionDenied
	}

	// 2. (Optional but Recommended) Validate that all provided track IDs exist
	if len(orderedTrackIDs) > 0 {
		existingTracks, err := uc.trackRepo.ListByIDs(ctx, orderedTrackIDs) // Fetch the tracks
		if err != nil {
			uc.logger.ErrorContext(ctx, "Failed to validate track IDs for collection update", "error", err, "collectionID", collectionID)
			return fmt.Errorf("failed to verify tracks: %w", err)
		}
		if len(existingTracks) != len(orderedTrackIDs) {
			// Find which IDs are missing
			providedSet := make(map[domain.TrackID]struct{}, len(orderedTrackIDs))
			for _, id := range orderedTrackIDs { providedSet[id] = struct{}{} }
			for _, track := range existingTracks { delete(providedSet, track.ID) }
			missingIDs := make([]string, 0, len(providedSet))
			for id := range providedSet { missingIDs = append(missingIDs, id.String()) }

			uc.logger.WarnContext(ctx, "Attempt to add non-existent tracks to collection", "collectionID", collectionID, "missingTrackIDs", missingIDs)
			return fmt.Errorf("%w: one or more track IDs do not exist: %v", domain.ErrInvalidArgument, missingIDs)
		}
	}


	// 3. Call repository method to manage tracks (handles transaction)
	err = uc.collectionRepo.ManageTracks(ctx, collectionID, orderedTrackIDs)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to manage tracks in repository", "error", err, "collectionID", collectionID, "userID", userID)
		return fmt.Errorf("failed to update collection tracks: %w", err) // Internal error
	}

	uc.logger.InfoContext(ctx, "Collection tracks updated", "collectionID", collectionID, "userID", userID, "trackCount", len(orderedTrackIDs))
	return nil
}

// DeleteCollection deletes a collection owned by the user.
func (uc *AudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok { return domain.ErrUnauthenticated }

	// 1. Verify ownership
	collection, err := uc.collectionRepo.FindByID(ctx, collectionID)
	if err != nil {
		return err // Handles ErrNotFound
	}
	if collection.OwnerID != userID {
		uc.logger.WarnContext(ctx, "Attempt to delete collection not owned by user", "collectionID", collectionID, "ownerID", collection.OwnerID, "userID", userID)
		return domain.ErrPermissionDenied
	}

	// 2. Delete from repository
	err = uc.collectionRepo.Delete(ctx, collectionID)
	if err != nil {
		// Handles ErrNotFound again just in case of race condition
		if !errors.Is(err, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to delete collection from repository", "error", err, "collectionID", collectionID, "userID", userID)
		}
		return err
	}

	uc.logger.InfoContext(ctx, "Collection deleted", "collectionID", collectionID, "userID", userID)
	return nil
}