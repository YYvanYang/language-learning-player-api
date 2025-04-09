// internal/usecase/audio_content_uc.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/yvanyang/language-learning-player-api/internal/adapter/handler/http/middleware"
	"github.com/yvanyang/language-learning-player-api/internal/config"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
	"github.com/yvanyang/language-learning-player-api/pkg/pagination"
)

// AudioContentUseCase handles business logic related to audio tracks and collections.
type AudioContentUseCase struct {
	trackRepo      port.AudioTrackRepository
	collectionRepo port.AudioCollectionRepository
	storageService port.FileStorageService
	txManager      port.TransactionManager
	// ADDED: Inject activity repo to fetch user specific data in GetAudioTrackDetails
	progressRepo  port.PlaybackProgressRepository
	bookmarkRepo  port.BookmarkRepository
	presignExpiry time.Duration
	cdnBaseURL    *url.URL
	logger        *slog.Logger
}

// NewAudioContentUseCase creates a new AudioContentUseCase.
// ADDED: progressRepo and bookmarkRepo dependencies
func NewAudioContentUseCase(
	cfg config.Config,
	tr port.AudioTrackRepository,
	cr port.AudioCollectionRepository,
	ss port.FileStorageService,
	tm port.TransactionManager,
	pr port.PlaybackProgressRepository, // Added
	br port.BookmarkRepository, // Added
	log *slog.Logger,
) *AudioContentUseCase {
	if tm == nil {
		log.Warn("AudioContentUseCase created without TransactionManager implementation. Transactional operations will fail.")
	}
	var parsedCdnBaseURL *url.URL
	var parseErr error
	if cfg.CDN.BaseURL != "" {
		parsedCdnBaseURL, parseErr = url.Parse(cfg.CDN.BaseURL)
		if parseErr != nil {
			log.Warn("Invalid CDN BaseURL in config, CDN rewriting disabled", "url", cfg.CDN.BaseURL, "error", parseErr)
			parsedCdnBaseURL = nil
		} else {
			log.Info("CDN Rewriting Enabled", "baseUrl", parsedCdnBaseURL.String())
		}
	}

	return &AudioContentUseCase{
		trackRepo:      tr,
		collectionRepo: cr,
		storageService: ss,
		txManager:      tm,
		progressRepo:   pr, // Added
		bookmarkRepo:   br, // Added
		presignExpiry:  cfg.Minio.PresignExpiry,
		cdnBaseURL:     parsedCdnBaseURL,
		logger:         log.With("usecase", "AudioContentUseCase"),
	}
}

// --- Track Use Cases ---

// Point 4: GetAudioTrackDetails retrieves details and user-specific info, returns result struct.
func (uc *AudioContentUseCase) GetAudioTrackDetails(ctx context.Context, trackID domain.TrackID) (*port.GetAudioTrackDetailsResult, error) {
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx) // Check if user is logged in

	track, err := uc.trackRepo.FindByID(ctx, trackID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			uc.logger.WarnContext(ctx, "Audio track not found", "trackID", trackID)
		} else {
			uc.logger.ErrorContext(ctx, "Failed to get audio track from repository", "error", err, "trackID", trackID)
		}
		return nil, err // Propagate error (NotFound or Internal)
	}

	// Authorization check (Example: only public tracks accessible anonymously)
	if !track.IsPublic && !userAuthenticated {
		uc.logger.WarnContext(ctx, "Anonymous user attempted to access private track", "trackID", trackID)
		return nil, domain.ErrUnauthenticated // Require login for private tracks
	}
	// Further checks (ownership, subscription) could go here if needed

	// Generate Presigned URL
	presignedURLStr, err := uc.storageService.GetPresignedGetURL(ctx, track.MinioBucket, track.MinioObjectKey, uc.presignExpiry)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to generate presigned URL for track", "error", err, "trackID", trackID)
		presignedURLStr = "" // Continue but log the error, URL will be empty
	}

	// Rewrite URL if CDN is configured
	finalPlayURL := uc.rewriteURLForCDN(ctx, presignedURLStr)

	result := &port.GetAudioTrackDetailsResult{
		Track:   track,
		PlayURL: finalPlayURL,
		// UserProgress and UserBookmarks will be filled below if user is authenticated
	}

	// Fetch user-specific data if authenticated
	if userAuthenticated {
		// Fetch Progress
		progress, errProg := uc.progressRepo.Find(ctx, userID, trackID)
		if errProg != nil && !errors.Is(errProg, domain.ErrNotFound) {
			uc.logger.ErrorContext(ctx, "Failed to get user progress for track details", "error", errProg, "trackID", trackID, "userID", userID)
			// Continue without progress, error is logged
		} else if errProg == nil {
			result.UserProgress = progress
		}

		// Fetch Bookmarks
		bookmarks, errBook := uc.bookmarkRepo.ListByUserAndTrack(ctx, userID, trackID)
		if errBook != nil {
			uc.logger.ErrorContext(ctx, "Failed to get user bookmarks for track details", "error", errBook, "trackID", trackID, "userID", userID)
			// Continue without bookmarks, error is logged
		} else {
			result.UserBookmarks = bookmarks // Assign even if empty slice
		}
	}

	uc.logger.InfoContext(ctx, "Successfully retrieved audio track details", "trackID", trackID, "authenticated", userAuthenticated)
	return result, nil
}

// rewriteURLForCDN is a helper to rewrite presigned URL if CDN is configured.
func (uc *AudioContentUseCase) rewriteURLForCDN(ctx context.Context, originalURL string) string {
	if uc.cdnBaseURL == nil || originalURL == "" {
		return originalURL // No CDN configured or no original URL
	}

	parsedOriginalURL, parseErr := url.Parse(originalURL)
	if parseErr != nil {
		uc.logger.ErrorContext(ctx, "Failed to parse original presigned URL for CDN rewriting", "url", originalURL, "error", parseErr)
		return originalURL // Fallback to original URL on parsing error
	}

	// Construct rewritten URL using CDN base and original path + query
	rewrittenURL := &url.URL{
		Scheme:   uc.cdnBaseURL.Scheme,
		Host:     uc.cdnBaseURL.Host,
		Path:     parsedOriginalURL.Path,
		RawQuery: parsedOriginalURL.RawQuery, // Preserve signature etc.
	}

	finalURL := rewrittenURL.String()
	uc.logger.DebugContext(ctx, "Rewrote presigned URL for CDN", "original", originalURL, "rewritten", finalURL)
	return finalURL
}

// Point 5: ListTracks now takes input port.ListTracksInput
func (uc *AudioContentUseCase) ListTracks(ctx context.Context, input port.ListTracksInput) ([]*domain.AudioTrack, int, pagination.Page, error) {
	pageParams := pagination.NewPageFromOffset(input.Page.Limit, input.Page.Offset)

	// Map Usecase input to Repository filters
	repoFilters := port.ListTracksFilters{
		Query:         input.Query,
		LanguageCode:  input.LanguageCode,
		Level:         input.Level,
		IsPublic:      input.IsPublic,
		UploaderID:    input.UploaderID,
		Tags:          input.Tags,
		SortBy:        input.SortBy,
		SortDirection: input.SortDirection,
	}

	tracks, total, err := uc.trackRepo.List(ctx, repoFilters, pageParams)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to list audio tracks from repository", "error", err, "filters", repoFilters, "page", pageParams)
		return nil, 0, pageParams, fmt.Errorf("failed to retrieve track list: %w", err)
	}
	uc.logger.InfoContext(ctx, "Successfully listed audio tracks", "count", len(tracks), "total", total, "input", input)
	return tracks, total, pageParams, nil
}

// --- Collection Use Cases --- (No changes needed for the requested points in these methods)

func (uc *AudioContentUseCase) CreateCollection(ctx context.Context, title, description string, colType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthenticated
	}
	if uc.txManager == nil {
		return nil, fmt.Errorf("internal configuration error: transaction manager not available")
	}

	collection, err := domain.NewAudioCollection(title, description, userID, colType)
	if err != nil {
		return nil, err
	}

	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		if err := uc.collectionRepo.Create(txCtx, collection); err != nil {
			return fmt.Errorf("saving collection metadata: %w", err)
		}
		if len(initialTrackIDs) > 0 {
			exists, validateErr := uc.validateTrackIDsExist(txCtx, initialTrackIDs)
			if validateErr != nil {
				return fmt.Errorf("validating initial tracks: %w", validateErr)
			}
			if !exists {
				return fmt.Errorf("%w: one or more initial track IDs do not exist", domain.ErrInvalidArgument)
			}
			if err := uc.collectionRepo.ManageTracks(txCtx, collection.ID, initialTrackIDs); err != nil {
				return fmt.Errorf("adding initial tracks: %w", err)
			}
			collection.TrackIDs = initialTrackIDs
		}
		return nil
	})

	if finalErr != nil {
		uc.logger.ErrorContext(ctx, "Transaction failed during collection creation", "error", finalErr, "collectionID", collection.ID, "userID", userID)
		return nil, fmt.Errorf("failed to create collection: %w", finalErr)
	}
	uc.logger.InfoContext(ctx, "Audio collection created", "collectionID", collection.ID, "userID", userID)
	return collection, nil
}

func (uc *AudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	userID, userAuthenticated := middleware.GetUserIDFromContext(ctx)
	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
	if err != nil { /* Handle NotFound, log other errors */
		return nil, err
	}
	if !userAuthenticated || collection.OwnerID != userID { /* Log, return permission denied */
		return nil, domain.ErrPermissionDenied
	}
	uc.logger.InfoContext(ctx, "Successfully retrieved collection details", "collectionID", collectionID, "trackCount", len(collection.TrackIDs))
	return collection, nil
}

func (uc *AudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
	collection, err := uc.collectionRepo.FindWithTracks(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	if len(collection.TrackIDs) == 0 {
		return []*domain.AudioTrack{}, nil
	}
	tracks, err := uc.trackRepo.ListByIDs(ctx, collection.TrackIDs)
	if err != nil { /* Log, return wrapped error */
		return nil, fmt.Errorf("failed to retrieve track details for collection: %w", err)
	}
	uc.logger.InfoContext(ctx, "Successfully retrieved tracks for collection", "collectionID", collectionID, "trackCount", len(tracks))
	return tracks, nil
}

func (uc *AudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title, description string) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return domain.ErrUnauthenticated
	}
	if title == "" {
		return fmt.Errorf("%w: collection title cannot be empty", domain.ErrInvalidArgument)
	}
	tempCollection := &domain.AudioCollection{ID: collectionID, OwnerID: userID, Title: title, Description: description}
	err := uc.collectionRepo.UpdateMetadata(ctx, tempCollection)
	if err != nil { /* Log if not NotFound/PermissionDenied, return err */
		return err
	}
	uc.logger.InfoContext(ctx, "Collection metadata updated", "collectionID", collectionID, "userID", userID)
	return nil
}

func (uc *AudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return domain.ErrUnauthenticated
	}
	if uc.txManager == nil {
		return fmt.Errorf("internal configuration error: transaction manager not available")
	}

	finalErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		collection, err := uc.collectionRepo.FindByID(txCtx, collectionID)
		if err != nil {
			return err
		}
		if collection.OwnerID != userID {
			return domain.ErrPermissionDenied
		}
		if len(orderedTrackIDs) > 0 {
			exists, validateErr := uc.validateTrackIDsExist(txCtx, orderedTrackIDs)
			if validateErr != nil {
				return fmt.Errorf("validating tracks: %w", validateErr)
			}
			if !exists {
				return fmt.Errorf("%w: one or more track IDs do not exist", domain.ErrInvalidArgument)
			}
		}
		if err := uc.collectionRepo.ManageTracks(txCtx, collectionID, orderedTrackIDs); err != nil {
			return fmt.Errorf("updating collection tracks in repository: %w", err)
		}
		return nil
	})

	if finalErr != nil { /* Log, return wrapped error */
		return fmt.Errorf("failed to update collection tracks: %w", finalErr)
	}
	uc.logger.InfoContext(ctx, "Collection tracks updated", "collectionID", collectionID, "userID", userID, "trackCount", len(orderedTrackIDs))
	return nil
}

func (uc *AudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return domain.ErrUnauthenticated
	}
	collection, err := uc.collectionRepo.FindByID(ctx, collectionID)
	if err != nil {
		return err
	}
	if collection.OwnerID != userID {
		return domain.ErrPermissionDenied
	}
	err = uc.collectionRepo.Delete(ctx, collectionID)
	if err != nil { /* Log if not NotFound, return err */
		return err
	}
	uc.logger.InfoContext(ctx, "Collection deleted", "collectionID", collectionID, "userID", userID)
	return nil
}

// Helper remains the same
func (uc *AudioContentUseCase) validateTrackIDsExist(ctx context.Context, trackIDs []domain.TrackID) (bool, error) {
	if len(trackIDs) == 0 {
		return true, nil
	}
	existingTracks, err := uc.trackRepo.ListByIDs(ctx, trackIDs)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to validate track IDs existence", "error", err)
		return false, fmt.Errorf("failed to verify tracks: %w", err)
	}
	if len(existingTracks) != len(trackIDs) { /* Log missing IDs */
		return false, nil
	}
	return true, nil
}

var _ port.AudioContentUseCase = (*AudioContentUseCase)(nil)
