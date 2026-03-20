package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type actualExpenseRepository struct{}

func NewActualExpenseRepository() ActualExpenseRepository {
	return &actualExpenseRepository{}
}

func (r *actualExpenseRepository) Create(ctx context.Context, tx pgx.Tx, expense *domain.ActualExpense, encryptedAmount string) error {
	query := `
		INSERT INTO actual_expenses (name, description, expense_date, encrypted_amount, budget_id, category_id, expected_expense_id, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	now := time.Now()
	expense.ExternalID = uuid.New()
	expense.CreatedAt = now
	expense.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		expense.Name,
		expense.Description,
		expense.ExpenseDate,
		encryptedAmount,
		expense.BudgetID,
		expense.CategoryID,
		expense.ExpectedExpenseID,
		expense.ExternalID,
		expense.CreatedAt,
		expense.UpdatedAt,
	).Scan(&expense.ID)
}

func (r *actualExpenseRepository) GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.ActualExpense, string, error) {
	query := `
		SELECT id, external_id, name, description, expense_date, encrypted_amount, budget_id, category_id, expected_expense_id, created_at, updated_at, revoked_at
		FROM actual_expenses
		WHERE external_id = $1 AND revoked_at IS NULL
	`

	expense := &domain.ActualExpense{}
	var encryptedAmount string
	err := tx.QueryRow(ctx, query, externalID).Scan(
		&expense.ID,
		&expense.ExternalID,
		&expense.Name,
		&expense.Description,
		&expense.ExpenseDate,
		&encryptedAmount,
		&expense.BudgetID,
		&expense.CategoryID,
		&expense.ExpectedExpenseID,
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

func (r *actualExpenseRepository) GetByBudgetID(ctx context.Context, tx pgx.Tx, budgetID int64) ([]domain.ActualExpense, []string, error) {
	query := `
		SELECT id, external_id, name, description, expense_date, encrypted_amount, budget_id, category_id, expected_expense_id, created_at, updated_at, revoked_at
		FROM actual_expenses
		WHERE budget_id = $1 AND revoked_at IS NULL
		ORDER BY expense_date DESC
	`

	rows, err := tx.Query(ctx, query, budgetID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var expenses []domain.ActualExpense
	var encryptedAmounts []string
	for rows.Next() {
		var e domain.ActualExpense
		var encryptedAmount string
		if err := rows.Scan(
			&e.ID,
			&e.ExternalID,
			&e.Name,
			&e.Description,
			&e.ExpenseDate,
			&encryptedAmount,
			&e.BudgetID,
			&e.CategoryID,
			&e.ExpectedExpenseID,
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

func (r *actualExpenseRepository) Update(ctx context.Context, tx pgx.Tx, expense *domain.ActualExpense, encryptedAmount string) error {
	query := `
		UPDATE actual_expenses
		SET name = $1, description = $2, expense_date = $3, encrypted_amount = $4, category_id = $5, expected_expense_id = $6, updated_at = $7
		WHERE id = $8 AND revoked_at IS NULL
	`

	expense.UpdatedAt = time.Now()

	result, err := tx.Exec(ctx, query,
		expense.Name,
		expense.Description,
		expense.ExpenseDate,
		encryptedAmount,
		expense.CategoryID,
		expense.ExpectedExpenseID,
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

func (r *actualExpenseRepository) Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error {
	query := `
		UPDATE actual_expenses
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
