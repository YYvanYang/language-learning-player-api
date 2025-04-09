// internal/usecase/user_uc.go
package usecase

import (
	"context"
	"log/slog"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

// userUseCase implements the port.UserUseCase interface.
type userUseCase struct {
	userRepo port.UserRepository
	logger   *slog.Logger
}

// NewUserUseCase creates a new UserUseCase.
func NewUserUseCase(ur port.UserRepository, log *slog.Logger) port.UserUseCase {
	return &userUseCase{
		userRepo: ur,
		logger:   log,
	}
}

// GetUserProfile retrieves a user's profile by their ID.
func (uc *userUseCase) GetUserProfile(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		// FindByID should return domain.ErrNotFound if not found
		uc.logger.WarnContext(ctx, "Failed to get user profile", "userID", userID, "error", err)
		return nil, err
	}
	// Note: We might want to selectively return fields or use a specific DTO
	// instead of the full domain.User if there's sensitive info like password hash.
	// However, the repository FindByID should already exclude the hash if necessary.
	return user, nil
}
