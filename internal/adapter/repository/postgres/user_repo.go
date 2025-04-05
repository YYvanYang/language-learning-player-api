// internal/adapter/repository/postgres/user_repo.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings" // Import strings package
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"   // Import pgconn for PgError
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-backend/internal/domain" // Adjust import path
	"github.com/yvanyang/language-learning-player-backend/internal/port"   // Adjust import path
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *pgxpool.Pool, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.With("repository", "UserRepository"), // Add context to logger
	}
}

// --- Interface Implementation ---

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email.String(), // Use string representation of value object
		user.Name,
		user.HashedPassword,
		user.GoogleID,
		user.AuthProvider,
		user.ProfileImageURL,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		// Check if the error is a PostgreSQL error and specifically a unique_violation
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			// Check which constraint was violated based on its name (defined in migration)
			// Common constraint names are like <table_name>_<column_name>_key
			if strings.Contains(pgErr.ConstraintName, "users_email_key") {
				// Specific error message for email conflict
				return fmt.Errorf("creating user: %w: email already exists", domain.ErrConflict)
			}
			if strings.Contains(pgErr.ConstraintName, "users_google_id_key") {
				// Specific error message for Google ID conflict
				return fmt.Errorf("creating user: %w: google ID already exists", domain.ErrConflict)
			}
			// Generic conflict if constraint name is unknown or not specifically handled
			r.logger.WarnContext(ctx, "Unique constraint violation on user creation", "constraint", pgErr.ConstraintName, "userID", user.ID)
			return fmt.Errorf("creating user: %w: resource conflict on unique field", domain.ErrConflict)
		}
		// If it's not a unique violation, log as internal error
		r.logger.ErrorContext(ctx, "Error creating user", "error", err, "userID", user.ID)
		return fmt.Errorf("creating user: %w", err)
	}
	r.logger.InfoContext(ctx, "User created successfully", "userID", user.ID, "email", user.Email.String())
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `
        SELECT id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at
        FROM users
        WHERE id = $1
    `
	user, err := r.scanUser(ctx, r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding user by ID", "error", err, "userID", id)
		return nil, fmt.Errorf("finding user by ID: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	query := `
        SELECT id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at
        FROM users
        WHERE email = $1
    `
	user, err := r.scanUser(ctx, r.db.QueryRow(ctx, query, email.String()))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding user by Email", "error", err, "email", email.String())
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerUserID string) (*domain.User, error) {
	// For now, only handling Google ID. Extend if more providers are added.
	if provider != domain.AuthProviderGoogle {
		return nil, fmt.Errorf("finding user by provider ID: provider '%s' not supported", provider)
	}

	query := `
        SELECT id, email, name, password_hash, google_id, auth_provider, profile_image_url, created_at, updated_at
        FROM users
        WHERE google_id = $1 AND auth_provider = $2
    `
	user, err := r.scanUser(ctx, r.db.QueryRow(ctx, query, providerUserID, provider))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding user by Provider ID", "error", err, "provider", provider, "providerUserID", providerUserID)
		return nil, fmt.Errorf("finding user by provider ID: %w", err)
	}
	return user, nil
}


func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	// Ensure updated_at is current
	user.UpdatedAt = time.Now()

	query := `
        UPDATE users
        SET email = $2, name = $3, password_hash = $4, google_id = $5, auth_provider = $6, profile_image_url = $7, updated_at = $8
        WHERE id = $1
    `
	cmdTag, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email.String(),
		user.Name,
		user.HashedPassword,
		user.GoogleID,
		user.AuthProvider,
		user.ProfileImageURL,
		user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		// Check for unique constraint violation on update (e.g., changing email to an existing one)
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolation {
			if strings.Contains(pgErr.ConstraintName, "users_email_key") {
				return fmt.Errorf("updating user: %w: email already exists", domain.ErrConflict)
			}
			if strings.Contains(pgErr.ConstraintName, "users_google_id_key") {
				return fmt.Errorf("updating user: %w: google ID already exists", domain.ErrConflict)
			}
			r.logger.WarnContext(ctx, "Unique constraint violation on user update", "constraint", pgErr.ConstraintName, "userID", user.ID)
			return fmt.Errorf("updating user: %w: resource conflict on unique field", domain.ErrConflict)
		}
		r.logger.ErrorContext(ctx, "Error updating user", "error", err, "userID", user.ID)
		return fmt.Errorf("updating user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		// If no rows were affected, the user ID likely didn't exist
		return domain.ErrNotFound // Or return a specific "update failed" error? ErrNotFound seems reasonable.
	}

	r.logger.InfoContext(ctx, "User updated successfully", "userID", user.ID)
	return nil
}


// --- Helper Methods ---

// scanUser is a helper function to scan a row into a domain.User object.
// It handles the conversion from SQL null types to Go pointers/value objects.
func (r *UserRepository) scanUser(ctx context.Context, row pgx.Row) (*domain.User, error) {
	var user domain.User
	var emailStr string // Scan email into a simple string first

	err := row.Scan(
		&user.ID,
		&emailStr,
		&user.Name,
		&user.HashedPassword, // Directly scans into *string (handles NULL)
		&user.GoogleID,       // Directly scans into *string (handles NULL)
		&user.AuthProvider,
		&user.ProfileImageURL, // Directly scans into *string (handles NULL)
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err // Let caller handle pgx.ErrNoRows or other errors
	}

	// Convert scanned email string to domain.Email value object
	emailVO, voErr := domain.NewEmail(emailStr)
	if voErr != nil {
		// This should ideally not happen if DB constraint is correct, but handle defensively
		r.logger.ErrorContext(ctx, "Invalid email format found in database", "error", voErr, "email", emailStr, "userID", user.ID)
		return nil, fmt.Errorf("invalid email format in DB for user %s: %w", user.ID, voErr)
	}
	user.Email = emailVO

	return &user, nil
}

// Compile-time check to ensure UserRepository satisfies the port.UserRepository interface
var _ port.UserRepository = (*UserRepository)(nil)