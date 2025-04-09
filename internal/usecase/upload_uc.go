// internal/usecase/upload_uc.go
package usecase

import (
	"context"
	"errors" // Import errors package
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
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
	logger         *slog.Logger
	minioBucket    string // Store default bucket name
}

// NewUploadUseCase creates a new UploadUseCase.
func NewUploadUseCase(
	cfg config.MinioConfig, // Need minio config for bucket name
	tr port.AudioTrackRepository,
	ss port.FileStorageService,
	log *slog.Logger,
) *UploadUseCase {
	return &UploadUseCase{
		trackRepo:      tr,
		storageService: ss,
		logger:         log.With("usecase", "UploadUseCase"),
		minioBucket:    cfg.BucketName,
	}
}

// RequestUploadResult holds the result of requesting an upload URL.
type RequestUploadResult struct {
	UploadURL string `json:"uploadUrl"`
	ObjectKey string `json:"objectKey"`
}

// RequestUpload generates a presigned PUT URL for the client to upload a file.
func (uc *UploadUseCase) RequestUpload(ctx context.Context, userID domain.UserID, filename string, contentType string) (*RequestUploadResult, error) {
	log := uc.logger.With("userID", userID.String(), "filename", filename, "contentType", contentType)

	// Basic validation (can add more, e.g., allowed content types, size limits if known)
	if filename == "" {
		return nil, fmt.Errorf("%w: filename cannot be empty", domain.ErrInvalidArgument)
	}
	// Optional: Validate contentType against allowed types
	allowedTypes := []string{"audio/mpeg", "audio/ogg", "audio/wav", "audio/aac", "audio/mp4"} // Example allowed types
	isAllowedType := false
	for _, t := range allowedTypes {
		if contentType == t {
			isAllowedType = true
			break
		}
	}
	if !isAllowedType {
		log.Warn("Upload requested with disallowed content type")
		// return nil, fmt.Errorf("%w: content type '%s' is not allowed", domain.ErrInvalidArgument, contentType)
		// Or allow it, but log warning. Let's allow for now.
	}

	// Generate a unique object key
	extension := filepath.Ext(filename) // Get extension (e.g., ".mp3")
	randomUUID := uuid.NewString()
	objectKey := fmt.Sprintf("user-uploads/%s/%s%s", userID.String(), randomUUID, extension) // Structure: prefix/userID/uuid.ext

	log = log.With("objectKey", objectKey) // Add objectKey to logger context

	// Get the presigned PUT URL from the storage service
	// Use a reasonable expiry time for the upload URL
	uploadURLExpiry := 15 * time.Minute // Configurable?
	uploadURL, err := uc.storageService.GetPresignedPutURL(ctx, uc.minioBucket, objectKey, uploadURLExpiry)
	if err != nil {
		log.Error("Failed to get presigned PUT URL", "error", err)
		return nil, fmt.Errorf("failed to prepare upload: %w", err) // Internal error
	}

	log.Info("Generated presigned URL for file upload")

	result := &RequestUploadResult{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
	}
	return result, nil
}

// CompleteUploadRequest holds the data needed to finalize an upload and create a track record.
type CompleteUploadRequest struct {
	ObjectKey     string
	Title         string
	Description   string
	LanguageCode  string
	Level         string // e.g., "A1", "B2"
	DurationMs    int64  // Duration in milliseconds (client needs to provide this, e.g., from HTML5 audio duration property)
	IsPublic      bool
	Tags          []string
	CoverImageURL *string
}

// CompleteUpload finalizes the upload process by creating an AudioTrack record in the database.
func (uc *UploadUseCase) CompleteUpload(ctx context.Context, userID domain.UserID, req CompleteUploadRequest) (*domain.AudioTrack, error) {
	log := uc.logger.With("userID", userID.String(), "objectKey", req.ObjectKey)

	// --- Validation ---
	if req.ObjectKey == "" {
		return nil, fmt.Errorf("%w: objectKey is required", domain.ErrInvalidArgument)
	}
	if req.Title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidArgument)
	}
	if req.LanguageCode == "" {
		return nil, fmt.Errorf("%w: languageCode is required", domain.ErrInvalidArgument)
	}
	if req.DurationMs <= 0 {
		return nil, fmt.Errorf("%w: valid durationMs is required", domain.ErrInvalidArgument)
	}

	// Validate object key prefix belongs to the user (basic security check)
	expectedPrefix := fmt.Sprintf("user-uploads/%s/", userID.String())
	if !strings.HasPrefix(req.ObjectKey, expectedPrefix) {
		log.Warn("Attempt to complete upload for object key not belonging to user", "expectedPrefix", expectedPrefix)
		// Return NotFound or PermissionDenied? NotFound might obscure the real issue.
		return nil, fmt.Errorf("%w: invalid object key provided", domain.ErrPermissionDenied)
	}

	// Optional: Check if object actually exists in MinIO?
	// This adds an extra call to MinIO but ensures the client *did* upload successfully.
	// _, err := uc.storageService.GetObjectInfo(ctx, uc.minioBucket, req.ObjectKey) // Assumes GetObjectInfo exists
	// if err != nil {
	//     log.Error("Failed to verify object existence in MinIO or access denied", "error", err)
	//     if errors.Is(err, domain.ErrNotFound) { // Assuming GetObjectInfo returns ErrNotFound
	//         return nil, fmt.Errorf("%w: uploaded file not found in storage", domain.ErrInvalidArgument)
	//     }
	//     return nil, fmt.Errorf("failed to verify upload: %w", err)
	// }

	// Validate Language and Level
	langVO, err := domain.NewLanguage(req.LanguageCode, "") // Name not stored
	if err != nil {
		return nil, err
	}
	levelVO := domain.AudioLevel(req.Level)
	if req.Level != "" && !levelVO.IsValid() { // Allow empty level
		return nil, fmt.Errorf("%w: invalid audio level '%s'", domain.ErrInvalidArgument, req.Level)
	}
	// --- End Validation ---

	// --- Create Domain Object ---
	duration := time.Duration(req.DurationMs) * time.Millisecond
	uploaderID := userID // Link track to the uploading user

	track, err := domain.NewAudioTrack(
		req.Title,
		req.Description,
		uc.minioBucket, // Use the configured bucket
		req.ObjectKey,
		langVO,
		levelVO,
		duration,
		&uploaderID,
		req.IsPublic,
		req.Tags,
		req.CoverImageURL,
	)
	if err != nil {
		log.Error("Failed to create AudioTrack domain object", "error", err, "request", req)
		return nil, fmt.Errorf("failed to process track data: %w", err) // Likely internal validation mismatch
	}
	// --- End Create Domain Object ---

	// --- Save to Repository ---
	err = uc.trackRepo.Create(ctx, track)
	if err != nil {
		log.Error("Failed to create audio track record in repository", "error", err, "trackID", track.ID)
		// Check for conflict (e.g., duplicate object key if somehow possible despite UUID)
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("%w: track with this identifier already exists", domain.ErrConflict)
		}
		// Handle other potential errors (e.g., database connection issues)
		return nil, fmt.Errorf("failed to save track information: %w", err) // Internal error
	}
	// --- End Save to Repository ---

	log.Info("Upload completed and track record created", "trackID", track.ID)
	return track, nil
}
