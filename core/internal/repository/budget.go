package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type budgetRepository struct{}

func NewBudgetRepository() BudgetRepository {
	return &budgetRepository{}
}

func (r *budgetRepository) Create(ctx context.Context, tx pgx.Tx, budget *domain.Budget) error {
	query := `
		INSERT INTO budgets (name, description, start_date, end_date, budgeting_group_id, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	budget.ExternalID = uuid.New()
	budget.CreatedAt = now
	budget.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		budget.Name,
		budget.Description,
		budget.StartDate,
		budget.EndDate,
		budget.BudgetingGroupID,
		budget.ExternalID,
		budget.CreatedAt,
		budget.UpdatedAt,
	).Scan(&budget.ID)
}

func (r *budgetRepository) GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.Budget, error) {
	query := `
		SELECT id, external_id, name, description, start_date, end_date, budgeting_group_id, created_at, updated_at, revoked_at
		FROM budgets
		WHERE external_id = $1 AND revoked_at IS NULL
	`

	budget := &domain.Budget{}
	err := tx.QueryRow(ctx, query, externalID).Scan(
		&budget.ID,
		&budget.ExternalID,
		&budget.Name,
		&budget.Description,
		&budget.StartDate,
		&budget.EndDate,
		&budget.BudgetingGroupID,
		&budget.CreatedAt,
		&budget.UpdatedAt,
		&budget.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return budget, nil
}

func (r *budgetRepository) GetByGroupID(ctx context.Context, tx pgx.Tx, groupID int64) ([]domain.Budget, error) {
	query := `
		SELECT id, external_id, name, description, start_date, end_date, budgeting_group_id, created_at, updated_at, revoked_at
		FROM budgets
		WHERE budgeting_group_id = $1 AND revoked_at IS NULL
		ORDER BY start_date DESC
	`

	rows, err := tx.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(
			&b.ID,
			&b.ExternalID,
			&b.Name,
			&b.Description,
			&b.StartDate,
			&b.EndDate,
			&b.BudgetingGroupID,
			&b.CreatedAt,
			&b.UpdatedAt,
			&b.RevokedAt,
		); err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}

	return budgets, rows.Err()
}

func (r *budgetRepository) Update(ctx context.Context, tx pgx.Tx, budget *domain.Budget) error {
	query := `
		UPDATE budgets
		SET name = $1, description = $2, start_date = $3, end_date = $4, updated_at = $5
		WHERE id = $6 AND revoked_at IS NULL
	`

	budget.UpdatedAt = time.Now()

	result, err := tx.Exec(ctx, query,
		budget.Name,
		budget.Description,
		budget.StartDate,
		budget.EndDate,
		budget.UpdatedAt,
		budget.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *budgetRepository) Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error {
	query := `
		UPDATE budgets
		SET revoked_at = $1, updated_at = $1
		WHERE external_id = $2 AND revoked_at IS NULL
	`

	now := time.Now()
	result, err := tx.Exec(ctx, query, now, externalID)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
