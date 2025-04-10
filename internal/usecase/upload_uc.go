// ============================================
// FILE: internal/usecase/upload_uc.go (MODIFIED)
// ============================================
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/yvanyang/language-learning-player-api/internal/config"
	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// UploadUseCase handles the business logic for file uploads.
type UploadUseCase struct {
	trackRepo      port.AudioTrackRepository
	storageService port.FileStorageService
	txManager      port.TransactionManager
	logger         *slog.Logger
	minioBucket    string
}

// NewUploadUseCase creates a new UploadUseCase.
func NewUploadUseCase(
	cfg config.MinioConfig,
	tr port.AudioTrackRepository,
	ss port.FileStorageService,
	tm port.TransactionManager,
	log *slog.Logger,
) *UploadUseCase {
	if tm == nil {
		log.Warn("UploadUseCase created without TransactionManager implementation. Batch completion will not be transactional.")
	}
	if ss == nil {
		log.Error("UploadUseCase created without FileStorageService implementation. Uploads will fail.")
	}
	return &UploadUseCase{
		trackRepo:      tr,
		storageService: ss,
		txManager:      tm,
		logger:         log.With("usecase", "UploadUseCase"),
		minioBucket:    cfg.BucketName,
	}
}

// RequestUpload generates a presigned PUT URL for the client to upload a single file.
func (uc *UploadUseCase) RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*port.RequestUploadResult, error) {
	log := uc.logger.With("userID", userID.String(), "filename", filename, "contentType", contentType)

	if filename == "" {
		return nil, fmt.Errorf("%w: filename cannot be empty", domain.ErrInvalidArgument)
	}
	_ = uc.validateContentType(contentType)

	objectKey := uc.generateObjectKey(userID, filename)
	log = log.With("objectKey", objectKey)

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}
	uploadURLExpiry := 15 * time.Minute
	uploadURL, err := uc.storageService.GetPresignedPutURL(ctx, uc.minioBucket, objectKey, uploadURLExpiry)
	if err != nil {
		log.Error("Failed to get presigned PUT URL", "error", err)
		return nil, fmt.Errorf("failed to prepare upload: %w", err)
	}

	log.Info("Generated presigned URL for file upload")
	result := &port.RequestUploadResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	}
	return result, nil
}

// CompleteUpload finalizes the upload process by creating an AudioTrack record in the database.
// CHANGED: Parameter type to port.CompleteUploadInput
func (uc *UploadUseCase) CompleteUpload(ctx context.Context, userID domain.UserID, input port.CompleteUploadInput) (*domain.AudioTrack, error) {
	log := uc.logger.With("userID", userID.String(), "objectKey", input.ObjectKey)

	// CHANGED: Use fields from input
	if err := uc.validateCompleteUploadRequest(ctx, userID, input.ObjectKey, input.Title, input.LanguageCode, input.Duration, input.Level); err != nil {
		return nil, err
	}

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}
	exists, checkErr := uc.storageService.ObjectExists(ctx, uc.minioBucket, input.ObjectKey)
	if checkErr != nil {
		log.Error("Failed to check object existence in storage", "error", checkErr)
		return nil, fmt.Errorf("failed to verify upload status: %w", checkErr)
	}
	if !exists {
		log.Warn("Attempted to complete upload for a non-existent object in storage")
		return nil, fmt.Errorf("%w: uploaded file not found in storage for the given key", domain.ErrInvalidArgument)
	}

	// CHANGED: Pass input struct
	track, err := uc.createDomainTrack(ctx, userID, input)
	if err != nil {
		return nil, err
	}

	err = uc.trackRepo.Create(ctx, track)
	if err != nil {
		log.Error("Failed to create audio track record in repository", "error", err, "trackID", track.ID)
		if errors.Is(err, domain.ErrConflict) {
			log.Warn("Conflict during track creation, potentially duplicate object key", "objectKey", input.ObjectKey)
			return nil, fmt.Errorf("%w: track identifier conflict, possibly duplicate object key", domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to save track information: %w", err) // Internal error
	}

	log.Info("Upload completed and track record created", "trackID", track.ID)
	return track, nil
}

// --- Batch Upload Methods ---

// RequestBatchUpload generates presigned PUT URLs for multiple files.
// CHANGED: Parameter type to port.BatchRequestUploadInput
func (uc *UploadUseCase) RequestBatchUpload(ctx context.Context, userID domain.UserID, input port.BatchRequestUploadInput) ([]port.BatchURLResultItem, error) {
	log := uc.logger.With("userID", userID.String(), "batchSize", len(input.Files))
	log.Info("Requesting batch upload URLs")

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}

	results := make([]port.BatchURLResultItem, len(input.Files))
	uploadURLExpiry := 15 * time.Minute

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, fileReq := range input.Files { // Iterate over input.Files
		wg.Add(1)
		go func(index int, f port.BatchRequestUploadInputItem) {
			defer wg.Done()
			itemLog := log.With("originalFilename", f.Filename, "contentType", f.ContentType)
			responseItem := port.BatchURLResultItem{
				OriginalFilename: f.Filename,
			}

			if f.Filename == "" {
				responseItem.Error = "filename cannot be empty"
			}
			_ = uc.validateContentType(f.ContentType)

			objectKey := uc.generateObjectKey(userID, f.Filename)
			responseItem.ObjectKey = objectKey
			itemLog = itemLog.With("objectKey", objectKey)

			if responseItem.Error == "" {
				uploadURL, err := uc.storageService.GetPresignedPutURL(ctx, uc.minioBucket, objectKey, uploadURLExpiry)
				if err != nil {
					itemLog.Error("Failed to get presigned PUT URL for batch item", "error", err)
					responseItem.Error = "failed to prepare upload URL"
				} else {
					responseItem.UploadURL = uploadURL
				}
			}

			mu.Lock()
			results[index] = responseItem
			mu.Unlock()
		}(i, fileReq)
	}

	wg.Wait()
	log.Info("Finished generating batch upload URLs")
	return results, nil
}

// CompleteBatchUpload finalizes multiple uploads within a single database transaction.
// CHANGED: Parameter type to port.BatchCompleteInput
func (uc *UploadUseCase) CompleteBatchUpload(ctx context.Context, userID domain.UserID, input port.BatchCompleteInput) ([]port.BatchCompleteResultItem, error) {
	log := uc.logger.With("userID", userID.String(), "batchSize", len(input.Tracks))
	log.Info("Attempting to complete batch upload")

	if uc.txManager == nil {
		return nil, fmt.Errorf("internal server error: batch processing misconfigured")
	}
	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}

	var processingErr error

	preCheckFailed := false
	validatedItems := make([]port.BatchCompleteItem, 0, len(input.Tracks))
	tempResults := make([]port.BatchCompleteResultItem, len(input.Tracks))

	for i, trackReq := range input.Tracks { // Iterate over input.Tracks
		itemLog := log.With("objectKey", trackReq.ObjectKey, "title", trackReq.Title)
		resultItem := port.BatchCompleteResultItem{
			ObjectKey: trackReq.ObjectKey,
			Success:   false,
		}

		validationErr := uc.validateCompleteUploadRequest(ctx, userID, trackReq.ObjectKey, trackReq.Title, trackReq.LanguageCode, trackReq.Duration, trackReq.Level)
		if validationErr != nil {
			itemLog.Warn("Pre-validation failed for batch item", "error", validationErr)
			resultItem.Error = validationErr.Error()
			preCheckFailed = true
		} else {
			exists, checkErr := uc.storageService.ObjectExists(ctx, uc.minioBucket, trackReq.ObjectKey)
			if checkErr != nil {
				itemLog.Error("Failed to check object existence pre-transaction", "error", checkErr)
				resultItem.Error = "failed to verify upload status"
				preCheckFailed = true
			} else if !exists {
				itemLog.Warn("Object not found in storage pre-transaction")
				resultItem.Error = "uploaded file not found in storage"
				preCheckFailed = true
			} else {
				validatedItems = append(validatedItems, trackReq)
				resultItem.Success = true
			}
		}
		tempResults[i] = resultItem
	}

	if preCheckFailed {
		log.Warn("Batch completion aborted due to pre-transaction validation/existence check failures")
		return tempResults, fmt.Errorf("%w: one or more items failed validation or were not found in storage", domain.ErrInvalidArgument)
	}

	log.Info("All items passed pre-checks, proceeding with database transaction.")
	finalDbResults := make(map[string]*port.BatchCompleteResultItem)
	for i := range tempResults {
		finalDbResults[tempResults[i].ObjectKey] = &tempResults[i]
	}

	txErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		var firstDbErr error
		for _, trackReq := range validatedItems {
			itemLog := log.With("objectKey", trackReq.ObjectKey, "title", trackReq.Title)
			resultItemPtr := finalDbResults[trackReq.ObjectKey]

			// CHANGED: Pass trackReq (port.BatchCompleteItem)
			track, domainErr := uc.createDomainTrack(txCtx, userID, trackReq)
			if domainErr != nil {
				itemLog.Error("Failed to create domain object for batch item (unexpected)", "error", domainErr)
				resultItemPtr.Success = false
				resultItemPtr.Error = "failed to process track data"
				if firstDbErr == nil {
					firstDbErr = fmt.Errorf("item %s failed: %w", trackReq.ObjectKey, domainErr)
				}
				continue
			}

			dbErr := uc.trackRepo.Create(txCtx, track)
			if dbErr != nil {
				itemLog.Error("Failed to create track record for batch item", "error", dbErr, "trackID", track.ID)
				resultItemPtr.Success = false
				resultItemPtr.Error = "failed to save track information"
				if errors.Is(dbErr, domain.ErrConflict) {
					resultItemPtr.Error = "track identifier conflict"
					log.Warn("Conflict during batch track creation, potentially duplicate object key", "objectKey", trackReq.ObjectKey)
				}
				if firstDbErr == nil {
					firstDbErr = fmt.Errorf("item %s failed: %w", trackReq.ObjectKey, dbErr)
				}
			} else {
				resultItemPtr.Success = true
				resultItemPtr.TrackID = track.ID.String()
				resultItemPtr.Error = ""
				itemLog.Info("Batch item processed and track created successfully in transaction", "trackID", track.ID)
			}
		}
		return firstDbErr
	})

	finalResults := make([]port.BatchCompleteResultItem, len(input.Tracks))
	for i := range tempResults {
		finalResults[i] = tempResults[i]
		if txErr != nil && finalDbResults[finalResults[i].ObjectKey] != nil && finalDbResults[finalResults[i].ObjectKey].Success {
			finalResults[i].Success = false
			if finalResults[i].Error == "" {
				finalResults[i].Error = "database transaction failed"
			}
		}
	}

	if txErr != nil {
		log.Error("Batch completion failed and transaction rolled back", "error", txErr)
		processingErr = fmt.Errorf("batch processing failed: %w", txErr)
	} else {
		log.Info("Batch completion finished and transaction committed")
	}

	return finalResults, processingErr
}

// --- Helper Methods ---

func (uc *UploadUseCase) validateContentType(contentType string) error {
	if contentType == "" {
		return fmt.Errorf("%w: contentType cannot be empty", domain.ErrInvalidArgument)
	}
	// Add more specific validation if needed (e.g., allow only audio/*)
	return nil
}

func (uc *UploadUseCase) generateObjectKey(userID domain.UserID, filename string) string {
	extension := filepath.Ext(filename)
	randomUUID := uuid.NewString()
	// Ensure consistent path separator
	return fmt.Sprintf("user-uploads/%s/%s%s", userID.String(), randomUUID, extension)
}

func (uc *UploadUseCase) validateCompleteUploadRequest(ctx context.Context, userID domain.UserID, objectKey, title, langCode string, duration time.Duration, level string) error {
	log := uc.logger.With("userID", userID.String(), "objectKey", objectKey)
	if objectKey == "" {
		return fmt.Errorf("%w: objectKey is required", domain.ErrInvalidArgument)
	}
	if title == "" {
		return fmt.Errorf("%w: title is required", domain.ErrInvalidArgument)
	}
	if langCode == "" {
		return fmt.Errorf("%w: languageCode is required", domain.ErrInvalidArgument)
	}
	if duration <= 0 {
		return fmt.Errorf("%w: valid duration is required", domain.ErrInvalidArgument)
	}
	expectedPrefix := fmt.Sprintf("user-uploads/%s/", userID.String())
	if !strings.HasPrefix(objectKey, expectedPrefix) {
		log.Warn("Attempt to complete upload for object key not belonging to user", "expectedPrefix", expectedPrefix)
		return fmt.Errorf("%w: invalid object key provided", domain.ErrPermissionDenied)
	}
	levelVO := domain.AudioLevel(level)
	if level != "" && !levelVO.IsValid() {
		return fmt.Errorf("%w: invalid audio level '%s'", domain.ErrInvalidArgument, level)
	}
	_, err := domain.NewLanguage(langCode, "")
	if err != nil {
		return err
	}
	return nil
}

// createDomainTrack creates an AudioTrack domain object from the request data.
func (uc *UploadUseCase) createDomainTrack(ctx context.Context, userID domain.UserID, reqData interface{}) (*domain.AudioTrack, error) {
	var title, description, objectKey, langCode, levelStr string
	var duration time.Duration
	var isPublic bool
	var tags []string
	var coverURL *string

	switch r := reqData.(type) {
	case port.CompleteUploadInput: // CHANGED: Use Input type
		title = r.Title
		description = r.Description
		objectKey = r.ObjectKey
		langCode = r.LanguageCode
		levelStr = r.Level
		duration = r.Duration
		isPublic = r.IsPublic
		tags = r.Tags
		coverURL = r.CoverImageURL
	case port.BatchCompleteItem:
		title = r.Title
		description = r.Description
		objectKey = r.ObjectKey
		langCode = r.LanguageCode
		levelStr = r.Level
		duration = r.Duration
		isPublic = r.IsPublic
		tags = r.Tags
		coverURL = r.CoverImageURL
	default:
		return nil, fmt.Errorf("internal error: unsupported type for createDomainTrack: %T", reqData)
	}

	langVO, err := domain.NewLanguage(langCode, "")
	if err != nil {
		return nil, err
	}
	levelVO := domain.AudioLevel(levelStr)
	uploaderID := userID

	track, err := domain.NewAudioTrack(title, description, uc.minioBucket, objectKey, langVO, levelVO, duration, &uploaderID, isPublic, tags, coverURL)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create AudioTrack domain object", "error", err)
		return nil, fmt.Errorf("failed to process track data: %w", err)
	}
	return track, nil
}

var _ port.UploadUseCase = (*UploadUseCase)(nil)
