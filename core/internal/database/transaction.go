package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxFunc func(ctx context.Context, tx pgx.Tx) error

func WithTransaction(ctx context.Context, pool *pgxpool.Pool, fn TxFunc) error {
	log.Printf("[DEBUG] WithTransaction: beginning transaction")
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("[DEBUG] WithTransaction: pool.Begin failed: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			log.Printf("[DEBUG] WithTransaction: panic recovered: %v, rolling back", p)
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		log.Printf("[DEBUG] WithTransaction: fn returned error: %v, rolling back", err)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Printf("[DEBUG] WithTransaction: rollback error: %v", rbErr)
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	log.Printf("[DEBUG] WithTransaction: fn returned nil, committing")
	if err := tx.Commit(ctx); err != nil {
		log.Printf("[DEBUG] WithTransaction: tx.Commit failed: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[DEBUG] WithTransaction: committed OK")
	return nil
}
