// internal/adapter/repository/postgres/tx_manager.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn" // ADDED: Import pgconn
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yvanyang/language-learning-player-api/internal/port" // Adjust import path
)

type txKey struct{} // Private key type for context value

// Querier defines the common methods needed from pgxpool.Pool and pgx.Tx
// Used by repositories to work transparently with or without a transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	// Add CopyFrom if needed
}

// RowScanner defines an interface for scanning database rows, compatible with pgx.Row and pgx.Rows.
// ADDED: Interface definition
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// getQuerier attempts to retrieve a pgx.Tx from the context.
// If not found, it returns the original *pgxpool.Pool.
// Both satisfy the Querier interface.
func getQuerier(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx // Use transaction if available
	}
	return pool // Fallback to pool
}

// TxManager implements the port.TransactionManager interface for PostgreSQL using pgx.
type TxManager struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewTxManager creates a new TxManager.
func NewTxManager(pool *pgxpool.Pool, logger *slog.Logger) *TxManager {
	return &TxManager{
		pool:   pool,
		logger: logger.With("component", "TxManager"),
	}
}

// Begin starts a new transaction and stores the pgx.Tx handle in the context.
func (tm *TxManager) Begin(ctx context.Context) (context.Context, error) {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		tm.logger.WarnContext(ctx, "Attempted to begin transaction within an existing transaction")
		return ctx, fmt.Errorf("transaction already in progress")
	}

	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		tm.logger.ErrorContext(ctx, "Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	tm.logger.DebugContext(ctx, "Transaction begun")
	txCtx := context.WithValue(ctx, txKey{}, tx)
	return txCtx, nil
}

// Commit commits the transaction stored in the context.
func (tm *TxManager) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		tm.logger.WarnContext(ctx, "Commit called without an active transaction in context")
		return fmt.Errorf("no transaction found in context to commit")
	}

	err := tx.Commit(ctx)
	if err != nil {
		tm.logger.ErrorContext(ctx, "Failed to commit transaction", "error", err)
		return fmt.Errorf("transaction commit failed: %w", err)
	}
	tm.logger.DebugContext(ctx, "Transaction committed")
	return nil
}

// Rollback rolls back the transaction stored in the context.
func (tm *TxManager) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		tm.logger.WarnContext(ctx, "Rollback called without an active transaction in context")
		return fmt.Errorf("no transaction found in context to rollback")
	}

	err := tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) { // Don't error if already closed
		tm.logger.ErrorContext(ctx, "Failed to rollback transaction", "error", err)
		return fmt.Errorf("transaction rollback failed: %w", err)
	}
	if err == nil {
		tm.logger.DebugContext(ctx, "Transaction rolled back")
	}
	return nil // Return nil even if ErrTxClosed
}

// Execute runs the given function within a transaction.
func (tm *TxManager) Execute(ctx context.Context, fn func(txCtx context.Context) error) (err error) {
	txCtx, beginErr := tm.Begin(ctx)
	if beginErr != nil {
		return beginErr
	}

	defer func() {
		if p := recover(); p != nil {
			rbErr := tm.Rollback(txCtx)
			err = fmt.Errorf("panic occurred during transaction: %v", p)
			if rbErr != nil {
				tm.logger.ErrorContext(ctx, "Failed to rollback after panic", "rollbackError", rbErr, "panic", p)
				err = fmt.Errorf("panic (%v) occurred and rollback failed: %w", p, rbErr)
			} else {
				tm.logger.ErrorContext(ctx, "Transaction rolled back due to panic", "panic", p)
			}
		}
	}()

	if fnErr := fn(txCtx); fnErr != nil {
		if rbErr := tm.Rollback(txCtx); rbErr != nil {
			tm.logger.ErrorContext(ctx, "Failed to rollback after function error", "rollbackError", rbErr, "functionError", fnErr)
			return fmt.Errorf("function failed (%w) and rollback failed: %w", fnErr, rbErr)
		}
		tm.logger.WarnContext(ctx, "Transaction rolled back due to function error", "error", fnErr)
		return fnErr // Return the original error from the function
	}

	if commitErr := tm.Commit(txCtx); commitErr != nil {
		tm.logger.ErrorContext(ctx, "Failed to commit transaction, attempting rollback", "commitError", commitErr)
		// Rollback might fail if the connection is broken after commit failure
		if rbErr := tm.Rollback(txCtx); rbErr != nil {
			tm.logger.ErrorContext(ctx, "Failed to rollback after commit failure", "rollbackError", rbErr, "commitError", commitErr)
			// Return both errors? Or just the commit error? Commit error is primary.
			return fmt.Errorf("commit failed (%w) and subsequent rollback failed: %w", commitErr, rbErr)
		}
		return commitErr // Return the commit error
	}

	tm.logger.DebugContext(ctx, "Transaction executed and committed successfully")
	return nil
}

// Compile-time check
var _ port.TransactionManager = (*TxManager)(nil)
