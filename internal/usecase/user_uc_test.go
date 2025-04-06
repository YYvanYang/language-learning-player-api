// =============================================
// FILE: internal/usecase/user_uc_test.go
// =============================================
package usecase_test

import (
	"context"
	"testing"
	"io"
	"log/slog"

	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port/mocks" // Use mocks
	"github.com/yvanyang/language-learning-player-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Re-use logger helper if available, otherwise define it
func newTestLoggerForUserUC() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestUserUseCase_GetUserProfile_Success(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository(t)
	logger := newTestLoggerForUserUC()
	uc := usecase.NewUserUseCase(mockUserRepo, logger)

	userID := domain.NewUserID()
	emailVO, _ := domain.NewEmail("test@example.com")
	expectedUser := &domain.User{
		ID:    userID,
		Email: emailVO,
		Name:  "Test User",
		// ... other fields ...
	}

	// Expect FindByID to be called and return the user
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(expectedUser, nil).Once()

	// Execute
	user, err := uc.GetUserProfile(context.Background(), userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, expectedUser, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUserProfile_NotFound(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository(t)
	logger := newTestLoggerForUserUC()
	uc := usecase.NewUserUseCase(mockUserRepo, logger)

	userID := domain.NewUserID()

	// Expect FindByID to return NotFound error
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(nil, domain.ErrNotFound).Once()

	// Execute
	user, err := uc.GetUserProfile(context.Background(), userID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUserProfile_RepoError(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository(t)
	logger := newTestLoggerForUserUC()
	uc := usecase.NewUserUseCase(mockUserRepo, logger)

	userID := domain.NewUserID()
	expectedError := errors.New("database connection error")

	// Expect FindByID to return a generic error
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(nil, expectedError).Once()

	// Execute
	user, err := uc.GetUserProfile(context.Background(), userID)

	// Assert
	require.Error(t, err)
	assert.Equal(t, expectedError, err) // Usecase should return the repo error directly
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}