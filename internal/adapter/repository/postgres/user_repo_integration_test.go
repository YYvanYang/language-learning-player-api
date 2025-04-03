// internal/adapter/repository/postgres/user_repo_integration_test.go
package postgres_test // Use _test package

import (
	"context"
	"testing"

	"your_project/internal/adapter/repository/postgres" // Import actual repo package
	"your_project/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a user repo instance for tests
func setupUserRepo(t *testing.T) *postgres.UserRepository {
    require.NotNil(t, testDBPool, "Database pool not initialized")
    require.NotNil(t, testLogger, "Logger not initialized")
    return postgres.NewUserRepository(testDBPool, testLogger)
}

// Helper to clean the users table before/after tests
func clearUsersTable(t *testing.T, ctx context.Context) {
    _, err := testDBPool.Exec(ctx, "DELETE FROM users")
    require.NoError(t, err, "Failed to clear users table")
}

func TestUserRepository_Integration_CreateAndFind(t *testing.T) {
    ctx := context.Background()
    repo := setupUserRepo(t)
    clearUsersTable(t, ctx) // Clean before test

    email := "integration@example.com"
    name := "Integration User"
    hashedPassword := "$2a$12$..." // Use a real (but maybe known) bcrypt hash for testing if needed, or generate one

    // Create user
    user, err := domain.NewLocalUser(email, name, hashedPassword)
    require.NoError(t, err)
    err = repo.Create(ctx, user)
    require.NoError(t, err, "Failed to create user")

    // Find by ID
    foundByID, err := repo.FindByID(ctx, user.ID)
    require.NoError(t, err, "Failed to find user by ID")
    require.NotNil(t, foundByID)
    assert.Equal(t, user.ID, foundByID.ID)
    assert.Equal(t, email, foundByID.Email.String())
    assert.Equal(t, name, foundByID.Name)
    require.NotNil(t, foundByID.HashedPassword)
    assert.Equal(t, hashedPassword, *foundByID.HashedPassword)
    assert.Equal(t, domain.AuthProviderLocal, foundByID.AuthProvider)

    // Find by Email
    emailVO, _ := domain.NewEmail(email)
    foundByEmail, err := repo.FindByEmail(ctx, emailVO)
    require.NoError(t, err, "Failed to find user by Email")
    require.NotNil(t, foundByEmail)
    assert.Equal(t, user.ID, foundByEmail.ID)

    // Find non-existent ID
    _, err = repo.FindByID(ctx, domain.NewUserID()) // Generate random ID
    require.Error(t, err, "Expected error for non-existent ID")
    assert.ErrorIs(t, err, domain.ErrNotFound, "Expected ErrNotFound")
}


func TestUserRepository_Integration_Create_ConflictEmail(t *testing.T) {
    ctx := context.Background()
    repo := setupUserRepo(t)
    clearUsersTable(t, ctx)

    email := "conflict@example.com"
    name1 := "User One"
    name2 := "User Two"
    hash := "somehash"

    // Create first user
    user1, _ := domain.NewLocalUser(email, name1, hash)
    err := repo.Create(ctx, user1)
    require.NoError(t, err)

    // Try creating second user with same email
    user2, _ := domain.NewLocalUser(email, name2, hash)
    err = repo.Create(ctx, user2)
    require.Error(t, err, "Expected error when creating user with duplicate email")
    assert.ErrorIs(t, err, domain.ErrConflict, "Expected ErrConflict for duplicate email")
    assert.Contains(t, err.Error(), "email already exists", "Error message should mention email conflict")
}

func TestUserRepository_Integration_Update(t *testing.T) {
    ctx := context.Background()
    repo := setupUserRepo(t)
    clearUsersTable(t, ctx)

    user, _ := domain.NewLocalUser("update@example.com", "Initial Name", "hash1")
    err := repo.Create(ctx, user)
    require.NoError(t, err)

    // Update fields
    user.Name = "Updated Name"
    newEmailStr := "updated@example.com"
    newEmailVO, _ := domain.NewEmail(newEmailStr)
    user.Email = newEmailVO
    user.GoogleID = nil // Ensure it updates NULL correctly

    err = repo.Update(ctx, user)
    require.NoError(t, err, "Failed to update user")

    // Verify update
    updatedUser, err := repo.FindByID(ctx, user.ID)
    require.NoError(t, err)
    assert.Equal(t, "Updated Name", updatedUser.Name)
    assert.Equal(t, newEmailStr, updatedUser.Email.String())
    assert.Nil(t, updatedUser.GoogleID)
    assert.True(t, updatedUser.UpdatedAt.After(user.CreatedAt), "UpdatedAt should be newer")
}

// TODO: Add tests for FindByProviderID
// TODO: Add tests for Update causing email conflict
// TODO: Add tests for deleting (if applicable)