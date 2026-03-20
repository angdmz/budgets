package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type expectedExpenseRepository struct{}

func NewExpectedExpenseRepository() ExpectedExpenseRepository {
	return &expectedExpenseRepository{}
}

func (r *expectedExpenseRepository) Create(ctx context.Context, tx pgx.Tx, expense *domain.ExpectedExpense, encryptedAmount string) error {
	query := `
		INSERT INTO expected_expenses (name, description, encrypted_amount, budget_id, category_id, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	expense.ExternalID = uuid.New()
	expense.CreatedAt = now
	expense.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		expense.Name,
		expense.Description,
		encryptedAmount,
		expense.BudgetID,
		expense.CategoryID,
		expense.ExternalID,
		expense.CreatedAt,
		expense.UpdatedAt,
	).Scan(&expense.ID)
}

func (r *expectedExpenseRepository) GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.ExpectedExpense, string, error) {
	query := `
		SELECT id, external_id, name, description, encrypted_amount, budget_id, category_id, created_at, updated_at, revoked_at
		FROM expected_expenses
		WHERE external_id = $1 AND revoked_at IS NULL
	`

	expense := &domain.ExpectedExpense{}
	var encryptedAmount string
	err := tx.QueryRow(ctx, query, externalID).Scan(
		&expense.ID,
		&expense.ExternalID,
		&expense.Name,
		&expense.Description,
		&encryptedAmount,
		&expense.BudgetID,
		&expense.CategoryID,
		&expense.CreatedAt,
		&expense.UpdatedAt,
		&expense.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, "", domain.ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}

	return expense, encryptedAmount, nil
}

func (r *expectedExpenseRepository) GetByBudgetID(ctx context.Context, tx pgx.Tx, budgetID int64) ([]domain.ExpectedExpense, []string, error) {
	query := `
		SELECT id, external_id, name, description, encrypted_amount, budget_id, category_id, created_at, updated_at, revoked_at
		FROM expected_expenses
		WHERE budget_id = $1 AND revoked_at IS NULL
		ORDER BY name
	`

	rows, err := tx.Query(ctx, query, budgetID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var expenses []domain.ExpectedExpense
	var encryptedAmounts []string
	for rows.Next() {
		var e domain.ExpectedExpense
		var encryptedAmount string
		if err := rows.Scan(
			&e.ID,
			&e.ExternalID,
			&e.Name,
			&e.Description,
			&encryptedAmount,
			&e.BudgetID,
			&e.CategoryID,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.RevokedAt,
		); err != nil {
			return nil, nil, err
		}
		expenses = append(expenses, e)
		encryptedAmounts = append(encryptedAmounts, encryptedAmount)
	}

	return expenses, encryptedAmounts, rows.Err()
}

func (r *expectedExpenseRepository) Update(ctx context.Context, tx pgx.Tx, expense *domain.ExpectedExpense, encryptedAmount string) error {
	query := `
		UPDATE expected_expenses
		SET name = $1, description = $2, encrypted_amount = $3, category_id = $4, updated_at = $5
		WHERE id = $6 AND revoked_at IS NULL
	`

	expense.UpdatedAt = time.Now()

	result, err := tx.Exec(ctx, query,
		expense.Name,
		expense.Description,
		encryptedAmount,
		expense.CategoryID,
		expense.UpdatedAt,
		expense.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *expectedExpenseRepository) Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error {
	query := `
		UPDATE expected_expenses
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
