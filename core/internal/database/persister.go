package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxPersister struct {
	tx pgx.Tx
}

func NewPgxPersister(tx pgx.Tx) *PgxPersister {
	return &PgxPersister{tx: tx}
}

func (p *PgxPersister) QueryRow(ctx context.Context, dest []any, query string, args ...any) error {
	return p.tx.QueryRow(ctx, query, args...).Scan(dest...)
}

func (p *PgxPersister) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	commandTag, err := p.tx.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return commandTag.RowsAffected(), nil
}

func (p *PgxPersister) QueryRows(ctx context.Context, dest func() []any, query string, args ...any) error {
	rows, err := p.tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		scanDest := dest()
		if err := rows.Scan(scanDest...); err != nil {
			return err
		}
	}

	return rows.Err()
}

func WithPersister(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context, p *PgxPersister) error) error {
	return WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		p := NewPgxPersister(tx)
		return fn(ctx, p)
	})
}
