// internal/usecase/upload_uc.go
package usecase

import (
	"context"
	"errors" // Import errors package
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync" // Import sync for WaitGroup
	"time"

	"github.com/google/uuid"

	// Assuming module path is now updated based on previous discussion
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
// Returns port.RequestUploadResult
func (uc *UploadUseCase) RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*port.RequestUploadResult, error) {
	log := uc.logger.With("userID", userID.String(), "filename", filename, "contentType", contentType)

	if filename == "" {
		return nil, fmt.Errorf("%w: filename cannot be empty", domain.ErrInvalidArgument)
	}
	_ = uc.validateContentType(contentType) // Validate but ignore error for now

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
	// Return the port-defined struct
	result := &port.RequestUploadResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	}
	return result, nil
}

// CompleteUpload finalizes the upload process by creating an AudioTrack record in the database.
// Accepts port.CompleteUploadParams (renamed from CompleteUploadRequest)
func (uc *UploadUseCase) CompleteUpload(ctx context.Context, userID domain.UserID, req port.CompleteUploadRequest) (*domain.AudioTrack, error) {
	log := uc.logger.With("userID", userID.String(), "objectKey", req.ObjectKey)

	// Validate using fields from port.CompleteUploadParams
	if err := uc.validateCompleteUploadRequest(ctx, userID, req.ObjectKey, req.Title, req.LanguageCode, req.DurationMs, req.Level); err != nil {
		return nil, err
	}

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}
	exists, checkErr := uc.storageService.ObjectExists(ctx, uc.minioBucket, req.ObjectKey)
	if checkErr != nil {
		log.Error("Failed to check object existence in storage", "error", checkErr)
		return nil, fmt.Errorf("failed to verify upload status: %w", checkErr)
	}
	if !exists {
		log.Warn("Attempted to complete upload for a non-existent object in storage")
		return nil, fmt.Errorf("%w: uploaded file not found in storage for the given key", domain.ErrInvalidArgument)
	}

	// Create domain track using fields from port.CompleteUploadParams
	track, err := uc.createDomainTrack(ctx, userID, req) // Pass the port struct
	if err != nil {
		return nil, err
	}

	err = uc.trackRepo.Create(ctx, track)
	if err != nil {
		log.Error("Failed to create audio track record in repository", "error", err, "trackID", track.ID)
		if errors.Is(err, domain.ErrConflict) {
			// Check if conflict is specifically due to object key
			// This requires a more specific error type from the repo or checking the error message.
			// For now, assume any conflict might be the object key.
			log.Warn("Conflict during track creation, potentially duplicate object key", "objectKey", req.ObjectKey)
			return nil, fmt.Errorf("%w: track identifier conflict, possibly duplicate object key", domain.ErrConflict)
		}
		return nil, fmt.Errorf("failed to save track information: %w", err) // Internal error
	}

	log.Info("Upload completed and track record created", "trackID", track.ID)
	return track, nil
}

// --- Batch Upload Methods ---

// RequestBatchUpload generates presigned PUT URLs for multiple files.
// Accepts port.BatchRequestUpload, returns []port.BatchURLResultItem
func (uc *UploadUseCase) RequestBatchUpload(ctx context.Context, userID domain.UserID, req port.BatchRequestUpload) ([]port.BatchURLResultItem, error) {
	log := uc.logger.With("userID", userID.String(), "batchSize", len(req.Files))
	log.Info("Requesting batch upload URLs")

	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}

	results := make([]port.BatchURLResultItem, len(req.Files))
	uploadURLExpiry := 15 * time.Minute

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, fileReq := range req.Files {
		wg.Add(1)
		go func(index int, f port.BatchRequestUploadItem) {
			defer wg.Done()
			itemLog := log.With("originalFilename", f.Filename, "contentType", f.ContentType)
			responseItem := port.BatchURLResultItem{
				OriginalFilename: f.Filename,
			}

			if f.Filename == "" {
				responseItem.Error = "filename cannot be empty"
			}
			_ = uc.validateContentType(f.ContentType) // Ignore error for now

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
// Accepts port.BatchCompleteRequest, returns []port.BatchCompleteResultItem
func (uc *UploadUseCase) CompleteBatchUpload(ctx context.Context, userID domain.UserID, req port.BatchCompleteRequest) ([]port.BatchCompleteResultItem, error) {
	log := uc.logger.With("userID", userID.String(), "batchSize", len(req.Tracks))
	log.Info("Attempting to complete batch upload")

	if uc.txManager == nil {
		return nil, fmt.Errorf("internal server error: batch processing misconfigured")
	}
	if uc.storageService == nil {
		return nil, fmt.Errorf("internal server error: storage service not available")
	}

	// results := make([]port.BatchCompleteResultItem, len(req.Tracks))
	var processingErr error

	preCheckFailed := false
	validatedItems := make([]port.BatchCompleteItem, 0, len(req.Tracks))
	tempResults := make([]port.BatchCompleteResultItem, len(req.Tracks)) // Use port type

	for i, trackReq := range req.Tracks { // Iterate over port.BatchCompleteItem
		itemLog := log.With("objectKey", trackReq.ObjectKey, "title", trackReq.Title)
		resultItem := port.BatchCompleteResultItem{ // Use port type
			ObjectKey: trackReq.ObjectKey,
			Success:   false,
		}

		validationErr := uc.validateCompleteUploadRequest(ctx, userID, trackReq.ObjectKey, trackReq.Title, trackReq.LanguageCode, trackReq.DurationMs, trackReq.Level)
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
				validatedItems = append(validatedItems, trackReq) // Add port.BatchCompleteItem
				resultItem.Success = true                         // Tentative
			}
		}
		tempResults[i] = resultItem
	}

	if preCheckFailed {
		log.Warn("Batch completion aborted due to pre-transaction validation/existence check failures")
		return tempResults, fmt.Errorf("%w: one or more items failed validation or were not found in storage", domain.ErrInvalidArgument)
	}

	log.Info("All items passed pre-checks, proceeding with database transaction.")
	finalDbResults := make(map[string]*port.BatchCompleteResultItem) // Map key to port type pointer
	for i := range tempResults {
		finalDbResults[tempResults[i].ObjectKey] = &tempResults[i]
	}

	txErr := uc.txManager.Execute(ctx, func(txCtx context.Context) error {
		var firstDbErr error
		for _, trackReq := range validatedItems { // Iterate over port.BatchCompleteItem
			itemLog := log.With("objectKey", trackReq.ObjectKey, "title", trackReq.Title)
			resultItemPtr := finalDbResults[trackReq.ObjectKey]

			track, domainErr := uc.createDomainTrack(txCtx, userID, trackReq) // Pass port.BatchCompleteItem
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

	finalResults := make([]port.BatchCompleteResultItem, len(req.Tracks)) // Use port type
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
	return nil
}

func (uc *UploadUseCase) generateObjectKey(userID domain.UserID, filename string) string {
	extension := filepath.Ext(filename)
	randomUUID := uuid.NewString()
	return fmt.Sprintf("user-uploads/%s/%s%s", userID.String(), randomUUID, extension)
}

func (uc *UploadUseCase) validateCompleteUploadRequest(ctx context.Context, userID domain.UserID, objectKey, title, langCode string, durationMs int64, level string) error {
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
	if durationMs <= 0 {
		return fmt.Errorf("%w: valid durationMs is required", domain.ErrInvalidArgument)
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
// Now accepts interface{} and performs type assertion for port structs.
func (uc *UploadUseCase) createDomainTrack(ctx context.Context, userID domain.UserID, reqData interface{}) (*domain.AudioTrack, error) {
	var title, description, objectKey, langCode, levelStr string
	var durationMs int64
	var isPublic bool
	var tags []string
	var coverURL *string

	// Type assertion for port structs
	switch r := reqData.(type) {
	case port.CompleteUploadRequest: // Single upload params from port
		title = r.Title
		description = r.Description
		objectKey = r.ObjectKey
		langCode = r.LanguageCode
		levelStr = r.Level
		durationMs = r.DurationMs
		isPublic = r.IsPublic
		tags = r.Tags
		coverURL = r.CoverImageURL
	case port.BatchCompleteItem: // Batch item from port
		title = r.Title
		description = r.Description
		objectKey = r.ObjectKey
		langCode = r.LanguageCode
		levelStr = r.Level
		durationMs = r.DurationMs
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
	duration := time.Duration(durationMs) * time.Millisecond
	uploaderID := userID

	track, err := domain.NewAudioTrack(title, description, uc.minioBucket, objectKey, langVO, levelVO, duration, &uploaderID, isPublic, tags, coverURL)
	if err != nil {
		uc.logger.ErrorContext(ctx, "Failed to create AudioTrack domain object", "error", err)
		return nil, fmt.Errorf("failed to process track data: %w", err)
	}
	return track, nil
}

// Compile-time check to ensure implementation satisfies the interface
var _ port.UploadUseCase = (*UploadUseCase)(nil)
