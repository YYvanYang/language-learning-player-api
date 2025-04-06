package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	httpadapter "github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http"
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/handler/http/dto"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port" // Import port for interfaces
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
	"github.com/yvanyang/language-learning-player-backend/pkg/httputil"
	"github.com/yvanyang/language-learning-player-backend/pkg/validation"
)

// --- Mock UseCase ---
// Use Mockery or manual mock based on port interface
type MockAudioContentUseCase struct {
	mock.Mock
	port.AudioContentUseCase // Embed the interface
}

// Implement only the methods needed for the test, forwarding calls to the mock object
func (m *MockAudioContentUseCase) ListTracks(ctx context.Context, params port.ListTracksParams, page port.Page) ([]*domain.AudioTrack, int, error) {
	args := m.Called(ctx, params, page)
	// Handle nil return for the slice pointer correctly
	if args.Get(0) == nil {
		// Return empty slice, total count, and error
		return []*domain.AudioTrack{}, args.Int(1), args.Error(2)
	}
	// Return tracks, total count, and error
	return args.Get(0).([]*domain.AudioTrack), args.Int(1), args.Error(2)
}

// Correct signatures based on audio_handler.go usage
func (m *MockAudioContentUseCase) GetTrackDetails(ctx context.Context, trackID domain.TrackID) (*domain.AudioTrack, string, error) {
	args := m.Called(ctx, trackID)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.AudioTrack), args.String(1), args.Error(2)
}

func (m *MockAudioContentUseCase) CreateCollection(ctx context.Context, title string, description string, collectionType domain.CollectionType, initialTrackIDs []domain.TrackID) (*domain.AudioCollection, error) {
	args := m.Called(ctx, title, description, collectionType, initialTrackIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AudioCollection), args.Error(1)
}

func (m *MockAudioContentUseCase) GetCollectionDetails(ctx context.Context, collectionID domain.CollectionID) (*domain.AudioCollection, error) {
	args := m.Called(ctx, collectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AudioCollection), args.Error(1)
}

func (m *MockAudioContentUseCase) GetCollectionTracks(ctx context.Context, collectionID domain.CollectionID) ([]*domain.AudioTrack, error) {
	args := m.Called(ctx, collectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AudioTrack), args.Error(1)
}

func (m *MockAudioContentUseCase) UpdateCollectionMetadata(ctx context.Context, collectionID domain.CollectionID, title string, description string) error {
	args := m.Called(ctx, collectionID, title, description)
	return args.Error(0)
}

func (m *MockAudioContentUseCase) DeleteCollection(ctx context.Context, collectionID domain.CollectionID) error {
	args := m.Called(ctx, collectionID)
	return args.Error(0)
}

func (m *MockAudioContentUseCase) UpdateCollectionTracks(ctx context.Context, collectionID domain.CollectionID, orderedTrackIDs []domain.TrackID) error {
	args := m.Called(ctx, collectionID, orderedTrackIDs)
	return args.Error(0)
}

// --- Test Function ---
func TestListTracks(t *testing.T) {
	validator := validation.New() // Use real validator

	// Define a sample domain object for use case response
	trackID, _ := domain.TrackIDFromString("uuid-track-1") // Use TrackIDFromString, ignore error for test setup
	sampleTrackDomain := &domain.AudioTrack{
		ID:            trackID,
		Title:         "Sample Track",
		Language:      "en-US",        // CHANGE: Correct field name
		Level:         "B1",
		Duration:      120 * time.Second, // CHANGE: Correct field name and type
		CoverImageURL: nil,
		IsPublic:      true,
		// Need MinioBucket and MinioObjectKey if MapDomainTrackToResponseDTO requires them
	}
	// Define the corresponding response DTO
	sampleTrackDTO := dto.AudioTrackResponseDTO{
		ID:            "uuid-track-1",
		Title:         "Sample Track",
		LanguageCode:  "en-US", // DTO field might still be LanguageCode
		Level:         "B1",
		DurationMs:    120000,  // DTO field might still be DurationMs
		CoverImageURL: nil,
		IsPublic:      true,
	}

	t.Run("Success - List tracks with pagination and filter", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase)
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)


		// Prepare expected response from use case (domain objects and total count)
		expectedTracks := []*domain.AudioTrack{sampleTrackDomain}
		expectedTotal := 1

		// Set expectations on the mock
		isPublicValue := true
		// Ensure pointers match the types in port.ListTracksParams
		langParam := string(sampleTrackDomain.Language) // Convert if necessary
		levelParam := domain.AudioLevel(sampleTrackDomain.Level)
		expectedParams := port.ListTracksParams{
			LanguageCode: &langParam, // CHANGE: Use pointer to string or correct type
			Level:        &levelParam,
			IsPublic:     &isPublicValue,
		}
		expectedPage := port.Page{Limit: 10, Offset: 0} // CHANGE: Use port.Page
		// Use mock.Anything for context, specific types for others
		mockUseCase.On("ListTracks", mock.Anything, expectedParams, expectedPage).Return(expectedTracks, expectedTotal, nil).Once()

		// Prepare HTTP request
		// Ensure query params match expectedParams fields
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks?lang=en-US&level=B1&limit=10&offset=0&isPublic=true", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusOK, rr.Code)

		var actualResponse dto.PaginatedTracksResponseDTO
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)

		// Assertions on paginated fields
		assert.Equal(t, expectedTotal, actualResponse.Total)
		assert.Equal(t, expectedPage.Limit, actualResponse.Limit)
		assert.Equal(t, expectedPage.Offset, actualResponse.Offset)
		assert.Len(t, actualResponse.Data, 1) // Check number of items

		// Optional: Deeper check of the actual data requires unmarshalling the interface{} slice items
		if len(actualResponse.Data) > 0 {
				// Need to handle the case where data is map[string]interface{} after json.Unmarshal
				var firstItemMap map[string]interface{}
				firstItemBytes, _ := json.Marshal(actualResponse.Data[0])
				err = json.Unmarshal(firstItemBytes, &firstItemMap)
				assert.NoError(t, err)

				// Compare map fields or remarshal to the expected DTO type
				var actualTrackDTO dto.AudioTrackResponseDTO
				err = json.Unmarshal(firstItemBytes, &actualTrackDTO)
				assert.NoError(t, err)
				assert.Equal(t, sampleTrackDTO, actualTrackDTO)
		}


		mockUseCase.AssertExpectations(t)
	})

	t.Run("Failure - UseCase returns error", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase)
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)

		expectedError := errors.New("internal database error")
		// Use mock.AnythingOfType for struct parameters if precise matching is difficult or not needed
		mockUseCase.On("ListTracks", mock.Anything, mock.AnythingOfType("port.ListTracksParams"), mock.AnythingOfType("port.Page")).Return(nil, 0, expectedError).Once() // CHANGE: Use port.Page

		// Prepare HTTP request
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var errResponse httputil.ErrorResponseDTO // CHANGE: Use correct DTO type
		err := json.Unmarshal(rr.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", errResponse.Code)
		// assert.Contains(t, errResponse.Message, "Failed to list audio tracks") // Check message if needed

		mockUseCase.AssertExpectations(t)
	})

	t.Run("Failure - Invalid query parameter value", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase) // UseCase should not be called
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)

		// Prepare HTTP request with invalid limit
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks?limit=abc", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusBadRequest, rr.Code) // Expect 400 due to bad input

		var errResponse httputil.ErrorResponseDTO // CHANGE: Use correct DTO type
		err := json.Unmarshal(rr.Body.Bytes(), &errResponse)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_INPUT", errResponse.Code)
		// assert.Contains(t, errResponse.Message, "Invalid value for parameter 'limit'") // Check message if needed

		// Assert that the use case method was NOT called
		mockUseCase.AssertNotCalled(t, "ListTracks", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Success - No query parameters (uses defaults)", func(t *testing.T) {
		// --- Arrange ---
		mockUseCase := new(MockAudioContentUseCase)
		var useCaseInterface port.AudioContentUseCase = mockUseCase
		handler := httpadapter.NewAudioHandler(useCaseInterface, validator)

		// Prepare expected response for default parameters
		expectedTracks := []*domain.AudioTrack{}
		expectedTotal := 0

		// Set expectations on the mock with default params
		// Adjust default values based on actual implementation in handler
		expectedParams := port.ListTracksParams{
			// Default IsPublic might be nil depending on handler logic
			IsPublic: nil, // Assuming nil means no filter by default
		}
		// Assuming default limit=20, offset=0 based on handler code
		expectedPage := port.Page{Limit: 20, Offset: 0} // CHANGE: Use port.Page
		mockUseCase.On("ListTracks", mock.Anything, expectedParams, expectedPage).Return(expectedTracks, expectedTotal, nil).Once()

		// Prepare HTTP request with no query params
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audio/tracks", nil)
		rr := httptest.NewRecorder()

		// --- Act ---
		handler.ListTracks(rr, req)

		// --- Assert ---
		assert.Equal(t, http.StatusOK, rr.Code)

		var actualResponse dto.PaginatedTracksResponseDTO
		err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
		assert.NoError(t, err)

		assert.Equal(t, expectedTotal, actualResponse.Total)
		assert.Equal(t, expectedPage.Limit, actualResponse.Limit)
		assert.Equal(t, expectedPage.Offset, actualResponse.Offset)
		assert.Len(t, actualResponse.Data, 0)

		mockUseCase.AssertExpectations(t)
	})

	// Add more test cases as needed.
}

// Ensure necessary DTOs and Param structs are defined:
// - internal/adapter/handler/http/dto/audio_dto.go needs AudioTrackResponseDTO and PaginatedTracksResponseDTO
// - internal/port/usecase.go needs AudioContentUseCase interface definition
// - internal/port/usecase.go or internal/port/repository.go needs ListTracksParams struct definition
// - internal/port/port.go (or similar) needs Page struct definition