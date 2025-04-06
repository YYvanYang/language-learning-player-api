// ==================================================
// FILE: internal/usecase/audio_content_uc_test.go
// ==================================================
package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/yvanyang/language-learning-player-backend/internal/config"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use mocks
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/middleware" // For context injection
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use logger helper if available
func newTestLoggerForAudioUC() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper to create a dummy MinIO config
func newTestMinioConfig() config.MinioConfig {
	return config.MinioConfig{BucketName: "test-bucket", PresignExpiry: 1 * time.Hour}
}

func TestAudioContentUseCase_GetAudioTrackDetails_Success(t *testing.T) {
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	mockCollectionRepo := mocks.NewMockAudioCollectionRepository(t) // Needed for constructor
	mockStorageService := mocks.NewMockFileStorageService(t)
    mockTxManager := mocks.NewMockTransactionManager(t) // Needed for constructor
	logger := newTestLoggerForAudioUC()
	cfg := newTestMinioConfig()

	uc := usecase.NewAudioContentUseCase(cfg, mockTrackRepo, mockCollectionRepo, mockStorageService, mockTxManager, logger)

	trackID := domain.NewTrackID()
	expectedTrack := &domain.AudioTrack{ID: trackID, Title: "Test", MinioBucket: cfg.BucketName, MinioObjectKey: "test/key"}
	expectedURL := "http://minio.example.com/presigned-url"

	// Mock expectations
	mockTrackRepo.On("FindByID", mock.Anything, trackID).Return(expectedTrack, nil).Once()
	mockStorageService.On("GetPresignedGetURL", mock.Anything, cfg.BucketName, expectedTrack.MinioObjectKey, cfg.PresignExpiry).Return(expectedURL, nil).Once()

	// Execute
	track, playURL, err := uc.GetAudioTrackDetails(context.Background(), trackID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedTrack, track)
	assert.Equal(t, expectedURL, playURL)
	mockTrackRepo.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
}

func TestAudioContentUseCase_GetAudioTrackDetails_TrackNotFound(t *testing.T) {
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	mockCollectionRepo := mocks.NewMockAudioCollectionRepository(t)
	mockStorageService := mocks.NewMockFileStorageService(t)
    mockTxManager := mocks.NewMockTransactionManager(t)
	logger := newTestLoggerForAudioUC()
	cfg := newTestMinioConfig()
	uc := usecase.NewAudioContentUseCase(cfg, mockTrackRepo, mockCollectionRepo, mockStorageService, mockTxManager, logger)

	trackID := domain.NewTrackID()

	// Expect FindByID to return NotFound
	mockTrackRepo.On("FindByID", mock.Anything, trackID).Return(nil, domain.ErrNotFound).Once()

	// Execute
	track, playURL, err := uc.GetAudioTrackDetails(context.Background(), trackID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, track)
	assert.Empty(t, playURL)
	mockTrackRepo.AssertExpectations(t)
	// Ensure storage service was not called
	mockStorageService.AssertNotCalled(t, "GetPresignedGetURL", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAudioContentUseCase_GetAudioTrackDetails_PresignError(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockCollectionRepo := mocks.NewMockAudioCollectionRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    mockTxManager := mocks.NewMockTransactionManager(t)
    logger := newTestLoggerForAudioUC()
    cfg := newTestMinioConfig()
    uc := usecase.NewAudioContentUseCase(cfg, mockTrackRepo, mockCollectionRepo, mockStorageService, mockTxManager, logger)

    trackID := domain.NewTrackID()
    expectedTrack := &domain.AudioTrack{ID: trackID, MinioBucket: cfg.BucketName, MinioObjectKey: "test/key"}
    presignError := errors.New("failed to connect to storage")

    // Mock expectations
    mockTrackRepo.On("FindByID", mock.Anything, trackID).Return(expectedTrack, nil).Once()
    mockStorageService.On("GetPresignedGetURL", mock.Anything, cfg.BucketName, expectedTrack.MinioObjectKey, cfg.PresignExpiry).Return("", presignError).Once()

    // Execute
    track, playURL, err := uc.GetAudioTrackDetails(context.Background(), trackID)

    // Assert
    require.Error(t, err)
    assert.Contains(t, err.Error(), "could not retrieve playback URL") // Check for wrapped error
    assert.Nil(t, track) // Should return nil track on presign failure
    assert.Empty(t, playURL)
    mockTrackRepo.AssertExpectations(t)
    mockStorageService.AssertExpectations(t)
}

func TestAudioContentUseCase_ListTracks_Success(t *testing.T) {
	mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
	mockCollectionRepo := mocks.NewMockAudioCollectionRepository(t)
	mockStorageService := mocks.NewMockFileStorageService(t)
    mockTxManager := mocks.NewMockTransactionManager(t)
	logger := newTestLoggerForAudioUC()
	cfg := newTestMinioConfig()
	uc := usecase.NewAudioContentUseCase(cfg, mockTrackRepo, mockCollectionRepo, mockStorageService, mockTxManager, logger)

	// Usecase input params
	ucParams := port.UseCaseListTracksParams{
		Page: pagination.Page{Limit: 10, Offset: 0}, // Requesting page 1, size 10
	}
	// Actual pagination after validation/defaults
	expectedPageParams := pagination.NewPageFromOffset(ucParams.Page.Limit, ucParams.Page.Offset)
	// Repo input params (assuming mapping for now)
	repoParams := port.ListTracksParams{}

	expectedTracks := []*domain.AudioTrack{{ID: domain.NewTrackID(), Title: "Track 1"}, {ID: domain.NewTrackID(), Title: "Track 2"}}
	expectedTotal := 25

	// Expect repo List to be called with correct parameters
	mockTrackRepo.On("List", mock.Anything, repoParams, expectedPageParams).Return(expectedTracks, expectedTotal, nil).Once()

	// Execute
	tracks, total, actualPage, err := uc.ListTracks(context.Background(), ucParams)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedTracks, tracks)
	assert.Equal(t, expectedTotal, total)
	assert.Equal(t, expectedPageParams, actualPage) // Check returned page info
	mockTrackRepo.AssertExpectations(t)
}

func TestAudioContentUseCase_ListTracks_RepoError(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockCollectionRepo := mocks.NewMockAudioCollectionRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    mockTxManager := mocks.NewMockTransactionManager(t)
    logger := newTestLoggerForAudioUC()
    cfg := newTestMinioConfig()
    uc := usecase.NewAudioContentUseCase(cfg, mockTrackRepo, mockCollectionRepo, mockStorageService, mockTxManager, logger)

    ucParams := port.UseCaseListTracksParams{ Page: pagination.Page{Limit: 10, Offset: 0} }
    expectedPageParams := pagination.NewPageFromOffset(ucParams.Page.Limit, ucParams.Page.Offset)
    repoParams := port.ListTracksParams{}
    repoError := errors.New("db connection lost")

    mockTrackRepo.On("List", mock.Anything, repoParams, expectedPageParams).Return(nil, 0, repoError).Once()

    // Execute
    tracks, total, actualPage, err := uc.ListTracks(context.Background(), ucParams)

    // Assert
    require.Error(t, err)
    assert.Contains(t, err.Error(), "failed to retrieve track list")
    assert.Nil(t, tracks)
    assert.Zero(t, total)
    assert.Equal(t, expectedPageParams, actualPage) // Should still return the page params used
    mockTrackRepo.AssertExpectations(t)
}

// --- Collection Tests ---

func TestAudioContentUseCase_CreateCollection_Success(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockCollectionRepo := mocks.NewMockAudioCollectionRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    mockTxManager := mocks.NewMockTransactionManager(t) // Use mock transaction manager
    logger := newTestLoggerForAudioUC()
    cfg := newTestMinioConfig()
    uc := usecase.NewAudioContentUseCase(cfg, mockTrackRepo, mockCollectionRepo, mockStorageService, mockTxManager, logger)

    userID := domain.NewUserID()
    ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID) // Add user ID to context
    title := "My Course"
    desc := "Course Desc"
    colType := domain.TypeCourse
    initialTrackIDs := []domain.TrackID{domain.NewTrackID(), domain.NewTrackID()}

    // Expect TransactionManager.Execute to be called
    // It will call the function passed to it. Inside that function:
    mockTxManager.On("Execute", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
        Run(func(args mock.Arguments) {
            // Execute the function passed to Execute, passing a context (doesn't matter which for mock repo)
            fn := args.Get(1).(func(context.Context) error)
            err := fn(ctx) // Pass the context, mock repo ignores tx value
            assert.NoError(t, err, "Function passed to Execute should succeed")
        }).
        Return(nil). // TransactionManager.Execute returns success
        Once()

    // Expect calls *within* the transaction function:
    // 1. CollectionRepo.Create
    var createdCollection *domain.AudioCollection
    mockCollectionRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.AudioCollection) bool {
        createdCollection = c // Capture created collection
        return c.Title == title && c.OwnerID == userID && c.Type == colType
    })).Return(nil).Once()
    // 2. TrackRepo.ListByIDs (for validation)
    mockTrackRepo.On("ListByIDs", mock.Anything, initialTrackIDs).Return([]*domain.AudioTrack{{}, {}}, nil).Once() // Assume tracks exist
    // 3. CollectionRepo.ManageTracks
    mockCollectionRepo.On("ManageTracks", mock.Anything, mock.AnythingOfType("domain.CollectionID"), initialTrackIDs).
        Run(func(args mock.Arguments) {
             require.NotNil(t, createdCollection, "Create must be called before ManageTracks")
             assert.Equal(t, createdCollection.ID, args.Get(1).(domain.CollectionID))
        }).
        Return(nil).Once()


    // Execute
    collection, err := uc.CreateCollection(ctx, title, desc, colType, initialTrackIDs)

    // Assert
    require.NoError(t, err)
    require.NotNil(t, collection)
    assert.Equal(t, title, collection.Title)
    assert.Equal(t, userID, collection.OwnerID)
    assert.Equal(t, initialTrackIDs, collection.TrackIDs) // Should be updated in memory

    // Verify mocks
    mockTxManager.AssertExpectations(t)
    mockCollectionRepo.AssertExpectations(t)
    mockTrackRepo.AssertExpectations(t)
}


func TestAudioContentUseCase_CreateCollection_TxFailureOnCreate(t *testing.T) {
    mockTrackRepo := mocks.NewMockAudioTrackRepository(t)
    mockCollectionRepo := mocks.NewMockAudioCollectionRepository(t)
    mockStorageService := mocks.NewMockFileStorageService(t)
    mockTxManager := mocks.NewMockTransactionManager(t)
    logger := newTestLoggerForAudioUC()
    cfg := newTestMinioConfig()
    uc := usecase.NewAudioContentUseCase(cfg, mockTrackRepo, mockCollectionRepo, mockStorageService, mockTxManager, logger)

    userID := domain.NewUserID()
    ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
    createError := errors.New("DB connection failed during create")

    // Expect TransactionManager.Execute
    mockTxManager.On("Execute", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
        Run(func(args mock.Arguments) {
            fn := args.Get(1).(func(context.Context) error)
            // Simulate error during the function execution (when Create is called)
            err := fn(ctx)
            assert.Error(t, err) // The function itself should return the error
            assert.Contains(t, err.Error(), "saving collection metadata")
        }).
        Return(fmt.Errorf("failed to create collection: %w", createError)). // TxManager Execute returns the wrapped error
        Once()

    // Expect CollectionRepo.Create to be called within Tx and return an error
    mockCollectionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AudioCollection")).Return(createError).Once()


    // Execute
    collection, err := uc.CreateCollection(ctx, "Title", "Desc", domain.TypePlaylist, nil)

    // Assert
    require.Error(t, err)
    assert.Contains(t, err.Error(), "failed to create collection") // Check the final wrapped error
    assert.ErrorContains(t, err, "DB connection failed during create") // Check original error is wrapped
    assert.Nil(t, collection)

    // Verify mocks
    mockTxManager.AssertExpectations(t)
    mockCollectionRepo.AssertExpectations(t)
    mockTrackRepo.AssertNotCalled(t, "ListByIDs") // Validation shouldn't happen if Create fails
    mockCollectionRepo.AssertNotCalled(t, "ManageTracks") // ManageTracks shouldn't happen
}

// TODO: Add TestCreateCollection_TxFailureOnManageTracks
// TODO: Add TestCreateCollection_TrackValidationFails
// TODO: Add TestGetCollectionDetails_Success
// TODO: Add TestGetCollectionDetails_NotFound
// TODO: Add TestGetCollectionDetails_PermissionDenied
// TODO: Add TestUpdateCollectionMetadata_Success
// TODO: Add TestUpdateCollectionMetadata_NotFound
// TODO: Add TestUpdateCollectionMetadata_PermissionDenied
// TODO: Add TestUpdateCollectionTracks_Success
// TODO: Add TestUpdateCollectionTracks_TxFailure
// TODO: Add TestUpdateCollectionTracks_TrackValidationFails
// TODO: Add TestDeleteCollection_Success
// TODO: Add TestDeleteCollection_NotFound
// TODO: Add TestDeleteCollection_PermissionDenied