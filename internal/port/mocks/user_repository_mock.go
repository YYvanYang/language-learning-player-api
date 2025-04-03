// internal/port/mocks/user_repository_mock.go
package mocks

import (
	"context"
	"your_project/internal/domain" // Adjust
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock type for the UserRepository type
type MockUserRepository struct {
	mock.Mock
}

// FindByID provides a mock function with given fields: ctx, id
func (_m *MockUserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	ret := _m.Called(ctx, id)
	var r0 *domain.User
	if rf, ok := ret.Get(0).(func(context.Context, domain.UserID) *domain.User); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.User)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, domain.UserID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// FindByEmail provides a mock function with given fields: ctx, email
func (_m *MockUserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	ret := _m.Called(ctx, email)
	// ... (similar pattern for return values) ...
	// Simplified:
	var r0 *domain.User
	if ret.Get(0) != nil { r0 = ret.Get(0).(*domain.User)}
	var r1 error = ret.Error(1)
	return r0, r1
}

// FindByProviderID provides a mock function with given fields: ctx, provider, providerUserID
func (_m *MockUserRepository) FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error) {
	ret := _m.Called(ctx, provider, providerUserID)
	// ... (similar pattern) ...
	var r0 *domain.User
	if ret.Get(0) != nil { r0 = ret.Get(0).(*domain.User)}
	var r1 error = ret.Error(1)
	return r0, r1
}

// Create provides a mock function with given fields: ctx, user
func (_m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	ret := _m.Called(ctx, user)
	var r0 error = ret.Error(0)
	return r0
}

// Update provides a mock function with given fields: ctx, user
func (_m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	ret := _m.Called(ctx, user)
	var r0 error = ret.Error(0)
	return r0
}

// Ensure mock satisfies the interface (optional but good practice)
// var _ port.UserRepository = (*MockUserRepository)(nil) // Cannot do this outside the original package