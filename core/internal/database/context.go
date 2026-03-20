package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type contextKey string

const txContextKey contextKey = "db_transaction"

// TxContext wraps a context with a database transaction
type TxContext struct {
	context.Context
	tx pgx.Tx
}

// ContextWithTransaction creates a new context with the given transaction
func ContextWithTransaction(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey, tx)
}

// TransactionFromContext retrieves the transaction from context
func TransactionFromContext(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txContextKey).(pgx.Tx); ok {
		return tx
	}
	return nil
}

// HasTransaction checks if the context has a transaction
func HasTransaction(ctx context.Context) bool {
	return TransactionFromContext(ctx) != nil
}
