// ==========================================================
// FILE: internal/adapter/repository/postgres/refreshtoken_repo.go (NEW FILE)
// ==========================================================
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-api/internal/domain"
	"github.com/yvanyang/language-learning-player-api/internal/port"
)

type RefreshTokenRepository struct {
	db         *pgxpool.Pool
	logger     *slog.Logger
	getQuerier func(ctx context.Context) Querier
}

func NewRefreshTokenRepository(db *pgxpool.Pool, logger *slog.Logger) *RefreshTokenRepository {
	repo := &RefreshTokenRepository{
		db:     db,
		logger: logger.With("repository", "RefreshTokenRepository"),
	}
	repo.getQuerier = func(ctx context.Context) Querier {
		return getQuerier(ctx, repo.db)
	}
	return repo
}

func (r *RefreshTokenRepository) Save(ctx context.Context, tokenData *port.RefreshTokenData) error {
	q := r.getQuerier(ctx)
	query := `
        INSERT INTO refresh_tokens (token_hash, user_id, expires_at, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (token_hash) DO UPDATE SET -- Should not happen if hashes are unique
            expires_at = EXCLUDED.expires_at,
            user_id = EXCLUDED.user_id -- Update user_id just in case (though unlikely to change)
    `
	if tokenData.CreatedAt.IsZero() {
		tokenData.CreatedAt = time.Now() // Ensure created_at is set
	}

	_, err := q.Exec(ctx, query,
		tokenData.TokenHash,
		tokenData.UserID,
		tokenData.ExpiresAt,
		tokenData.CreatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error saving refresh token", "error", err, "userID", tokenData.UserID)
		// Consider checking for FK violation on user_id -> domain.ErrInvalidArgument ?
		return fmt.Errorf("saving refresh token: %w", err)
	}
	r.logger.DebugContext(ctx, "Refresh token saved", "userID", tokenData.UserID, "expiresAt", tokenData.ExpiresAt)
	return nil
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*port.RefreshTokenData, error) {
	q := r.getQuerier(ctx)
	query := `
        SELECT token_hash, user_id, expires_at, created_at
        FROM refresh_tokens
        WHERE token_hash = $1
    `
	row := q.QueryRow(ctx, query, tokenHash)
	tokenData, err := r.scanTokenData(ctx, row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound // Map DB error to domain error
		}
		r.logger.ErrorContext(ctx, "Error finding refresh token by hash", "error", err)
		return nil, fmt.Errorf("finding refresh token by hash: %w", err)
	}
	return tokenData, nil
}

func (r *RefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	q := r.getQuerier(ctx)
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	cmdTag, err := q.Exec(ctx, query, tokenHash)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting refresh token by hash", "error", err)
		return fmt.Errorf("deleting refresh token by hash: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		// Not finding the token to delete is not necessarily an error in a logout/refresh flow
		r.logger.DebugContext(ctx, "Refresh token hash not found for deletion", "tokenHash", tokenHash) // Log hash prefix? No.
		return domain.ErrNotFound                                                                       // Or return nil? ErrNotFound seems slightly more informative.
	}
	r.logger.DebugContext(ctx, "Refresh token deleted by hash")
	return nil
}

func (r *RefreshTokenRepository) DeleteByUser(ctx context.Context, userID domain.UserID) (int64, error) {
	q := r.getQuerier(ctx)
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	cmdTag, err := q.Exec(ctx, query, userID)
	if err != nil {
		r.logger.ErrorContext(ctx, "Error deleting refresh tokens by user ID", "error", err, "userID", userID)
		return 0, fmt.Errorf("deleting refresh tokens by user ID: %w", err)
	}
	deletedCount := cmdTag.RowsAffected()
	r.logger.InfoContext(ctx, "Refresh tokens deleted by user ID", "userID", userID, "count", deletedCount)
	return deletedCount, nil
}

func (r *RefreshTokenRepository) scanTokenData(ctx context.Context, row RowScanner) (*port.RefreshTokenData, error) {
	var data port.RefreshTokenData
	err := row.Scan(
		&data.TokenHash,
		&data.UserID,
		&data.ExpiresAt,
		&data.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

var _ port.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
