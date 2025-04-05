// internal/usecase/mocks/auth_uc_mock.go
package mocks

import (
	"context"
	
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockAuthUseCase is a mock implementation of the AuthUseCase interface
type MockAuthUseCase struct {
	mock.Mock
}

// NewMockAuthUseCase creates a new MockAuthUseCase
func NewMockAuthUseCase(t interface{}) *MockAuthUseCase {
	return &MockAuthUseCase{}
}

// RegisterWithPassword is a mock implementation of the RegisterWithPassword method
func (m *MockAuthUseCase) RegisterWithPassword(ctx context.Context, emailStr, password, name string) (*domain.User, string, error) {
	args := m.Called(ctx, emailStr, password, name)
	
	var user *domain.User
	if args.Get(0) != nil {
		user = args.Get(0).(*domain.User)
	}
	
	return user, args.String(1), args.Error(2)
}

// LoginWithPassword is a mock implementation of the LoginWithPassword method
func (m *MockAuthUseCase) LoginWithPassword(ctx context.Context, emailStr, password string) (string, error) {
	args := m.Called(ctx, emailStr, password)
	return args.String(0), args.Error(1)
}

// AuthenticateWithGoogle is a mock implementation of the AuthenticateWithGoogle method
func (m *MockAuthUseCase) AuthenticateWithGoogle(ctx context.Context, googleIDToken string) (string, bool, error) {
	args := m.Called(ctx, googleIDToken)
	return args.String(0), args.Bool(1), args.Error(2)
} 