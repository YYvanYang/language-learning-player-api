// =============================================
// FILE: internal/usecase/upload_uc_test.go
// =============================================
package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use mocks
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use logger helper if available
func newTestLoggerForUploadUC() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Re-use MinIO config helper if available
func newTestMinioConfigForUpload() config.MinioConfig {
	return config.MinioConfig{BucketName: "upload-bucket", PresignExpiry: 15 * time.Minute}
}

func TestUploadUseCase_RequestUpload_Success(t *testing.T) {
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t) // Needed for constructor
	mockStorageService := mocks.NewMockFileStorageService(t)
	logger := newTestLoggerForUploadUC()
	cfg := newTestMinioConfigForUpload()

	uc := usecase.NewUploadUseCase(cfg, mockTrackRepo, mockStorageService, logger)

	userID := domain.NewUserID()
	filename := "podcast_episode.mp3"
	contentType := "audio/mpeg"
	expectedURL := "http://minio.example.com/presigned-put-url"

	// Expect storage service to be called to generate PUT URL
	mockStorageService.On("GetPresignedPutURL",
		mock.Anything, // Context
		cfg.BucketName,
		mock.MatchedBy(func(key string) bool { // Check the generated key format
			return strings.HasPrefix(key, fmt.Sprintf("user-uploads/%s/", userID.String())) &&
				strings.HasSuffix(key, ".mp3")
		}),
		15*time.Minute, // Expect the expiry used internally
	).Return(expectedURL, nil).Once()

	// Execute
	result, err := uc.RequestUpload(context.Background(), userID, filename, contentType)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedURL, result.UploadURL)
	assert.NotEmpty(t, result.ObjectKey)
	assert.True(t, strings.HasPrefix(result.ObjectKey, fmt.Sprintf("user-uploads/%s/", userID.String())))
	mockStorageService.AssertExpectations(t)
}

func TestUploadUseCase_RequestUpload_StorageError(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    logger := newTestLoggerForUploadUC()
    cfg := newTestMinioConfigForUpload()
    uc := usecase.NewUploadUseCase(cfg, mockTrackRepo, mockStorageService, logger)

    userID := domain.NewUserID()
    filename := "podcast_episode.mp3"
    contentType := "audio/mpeg"
    storageError := errors.New("minio connection refused")

    mockStorageService.On("GetPresignedPutURL", mock.Anything, cfg.BucketName, mock.AnythingOfType("string"), 15*time.Minute).
        Return("", storageError).Once()

    // Execute
    result, err := uc.RequestUpload(context.Background(), userID, filename, contentType)

    // Assert
    require.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "failed to prepare upload")
    assert.ErrorIs(t, err, storageError) // Check underlying error
    mockStorageService.AssertExpectations(t)
}

func TestUploadUseCase_RequestUpload_InvalidFilename(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    logger := newTestLoggerForUploadUC()
    cfg := newTestMinioConfigForUpload()
    uc := usecase.NewUploadUseCase(cfg, mockTrackRepo, mockStorageService, logger)

    userID := domain.NewUserID()
    filename := "" // Empty filename
    contentType := "audio/mpeg"

    // Execute
    result, err := uc.RequestUpload(context.Background(), userID, filename, contentType)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrInvalidArgument)
    assert.Nil(t, result)
    // Ensure storage service was NOT called
    mockStorageService.AssertNotCalled(t, "GetPresignedPutURL", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}


func TestUploadUseCase_CompleteUpload_Success(t *testing.T) {
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	mockStorageService := mocks.NewMockFileStorageService(t) // Needed for constructor
	logger := newTestLoggerForUploadUC()
	cfg := newTestMinioConfigForUpload()
	uc := usecase.NewUploadUseCase(cfg, mockTrackRepo, mockStorageService, logger)

	userID := domain.NewUserID()
	objectKey := fmt.Sprintf("user-uploads/%s/some-uuid.mp3", userID.String())
	req := usecase.CompleteUploadRequest{
		ObjectKey:    objectKey,
		Title:        "Uploaded Track",
		LanguageCode: "fr-FR",
		DurationMs:   185000,
		IsPublic:     true,
		Level:        "B2",
		Tags:         []string{"conversation"},
	}
    langVO, _ := domain.NewLanguage(req.LanguageCode, "")
    levelVO := domain.AudioLevel(req.Level)

	// Expect TrackRepo.Create to be called successfully
	var createdTrack *domain.AudioTrack
	mockTrackRepo.On("Create", mock.Anything, mock.MatchedBy(func(track *domain.AudioTrack) bool {
		createdTrack = track // Capture created track
		return track.MinioObjectKey == req.ObjectKey &&
			track.Title == req.Title &&
			track.Language == langVO &&
			track.Level == levelVO &&
			track.Duration == time.Duration(req.DurationMs)*time.Millisecond &&
			track.UploaderID != nil && *track.UploaderID == userID &&
			track.IsPublic == req.IsPublic &&
            track.MinioBucket == cfg.BucketName // Check bucket name
	})).Return(nil).Once()

	// Execute
	track, err := uc.CompleteUpload(context.Background(), userID, req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, track)
	assert.Equal(t, createdTrack.ID, track.ID) // Check if the returned track matches captured one
    assert.Equal(t, req.Title, track.Title)
	mockTrackRepo.AssertExpectations(t)
}

func TestUploadUseCase_CompleteUpload_InvalidObjectKeyPrefix(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    logger := newTestLoggerForUploadUC()
    cfg := newTestMinioConfigForUpload()
    uc := usecase.NewUploadUseCase(cfg, mockTrackRepo, mockStorageService, logger)

    userID := domain.NewUserID()
    otherUserID := domain.NewUserID()
    objectKey := fmt.Sprintf("user-uploads/%s/some-uuid.mp3", otherUserID.String()) // Belongs to other user
    req := usecase.CompleteUploadRequest{ ObjectKey: objectKey, Title: "Test", LanguageCode: "en", DurationMs: 1000 }

    // Execute
    track, err := uc.CompleteUpload(context.Background(), userID, req)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrPermissionDenied) // Expect permission denied due to key mismatch
    assert.Nil(t, track)
    mockTrackRepo.AssertNotCalled(t, "Create") // Ensure repo Create was not called
}


func TestUploadUseCase_CompleteUpload_RepoCreateConflict(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    logger := newTestLoggerForUploadUC()
    cfg := newTestMinioConfigForUpload()
    uc := usecase.NewUploadUseCase(cfg, mockTrackRepo, mockStorageService, logger)

    userID := domain.NewUserID()
    objectKey := fmt.Sprintf("user-uploads/%s/duplicate.mp3", userID.String())
    req := usecase.CompleteUploadRequest{ ObjectKey: objectKey, Title: "Test", LanguageCode: "en", DurationMs: 1000 }
    conflictError := fmt.Errorf("%w: duplicate key", domain.ErrConflict)

    // Expect TrackRepo.Create to return a conflict error
    mockTrackRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AudioTrack")).Return(conflictError).Once()

    // Execute
    track, err := uc.CompleteUpload(context.Background(), userID, req)

    // Assert
    require.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrConflict) // Usecase should propagate the conflict error
    assert.Nil(t, track)
    mockTrackRepo.AssertExpectations(t)
}

func TestUploadUseCase_CompleteUpload_InvalidInput(t *testing.T) {
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	mockStorageService := mocks.NewMockFileStorageService(t)
	logger := newTestLoggerForUploadUC()
	cfg := newTestMinioConfigForUpload()
	uc := usecase.NewUploadUseCase(cfg, mockTrackRepo, mockStorageService, logger)
	userID := domain.NewUserID()
	objectKey := fmt.Sprintf("user-uploads/%s/valid-key.mp3", userID)

	testCases := []struct {
		name string
		req  usecase.CompleteUploadRequest
		errSubstring string
	}{
		{"Missing ObjectKey", usecase.CompleteUploadRequest{Title:"T", LanguageCode:"en", DurationMs:1}, "objectKey is required"},
		{"Missing Title", usecase.CompleteUploadRequest{ObjectKey: objectKey, LanguageCode:"en", DurationMs:1}, "title is required"},
		{"Missing LangCode", usecase.CompleteUploadRequest{ObjectKey: objectKey, Title:"T", DurationMs:1}, "languageCode is required"},
		{"Zero Duration", usecase.CompleteUploadRequest{ObjectKey: objectKey, Title:"T", LanguageCode:"en", DurationMs:0}, "valid durationMs is required"},
		{"Invalid Level", usecase.CompleteUploadRequest{ObjectKey: objectKey, Title:"T", LanguageCode:"en", DurationMs:1, Level:"Z9"}, "invalid audio level"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.CompleteUpload(context.Background(), userID, tc.req)
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrInvalidArgument)
			assert.Contains(t, err.Error(), tc.errSubstring)
			mockTrackRepo.AssertNotCalled(t, "Create")
		})
	}
}