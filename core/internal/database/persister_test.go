package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/budgets/core/internal/domain"
)

// stubTx satisfies pgx.Tx; only QueryRow is exercised by PgxPersister.QueryRow.
type stubTx struct {
	pgx.Tx
	scanErr error
}

func (s stubTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return stubRow{err: s.scanErr}
}

type stubRow struct {
	err error
}

func (r stubRow) Scan(_ ...any) error { return r.err }

// Regression: pgx.ErrNoRows must be translated to domain.ErrNotFound so
// callers can distinguish missing rows from real failures (handoff Phase 0,
// item 2).
func TestPgxPersister_QueryRow_NoRowsTranslatesToNotFound(t *testing.T) {
	p := NewPgxPersister(stubTx{scanErr: pgx.ErrNoRows})

	var id int64
	err := p.QueryRow(context.Background(), []any{&id}, "SELECT 1")

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// Regression: non-NotFound errors must propagate unmasked.
func TestPgxPersister_QueryRow_OtherErrorsPropagate(t *testing.T) {
	boom := errors.New("connection reset")
	p := NewPgxPersister(stubTx{scanErr: boom})

	var id int64
	err := p.QueryRow(context.Background(), []any{&id}, "SELECT 1")

	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
	assert.NotErrorIs(t, err, domain.ErrNotFound)
}

func TestPgxPersister_QueryRow_Success(t *testing.T) {
	p := NewPgxPersister(stubTx{})

	var id int64
	err := p.QueryRow(context.Background(), []any{&id}, "SELECT 1")

	assert.NoError(t, err)
}
