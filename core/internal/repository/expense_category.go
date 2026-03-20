package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type expenseCategoryRepository struct{}

func NewExpenseCategoryRepository() ExpenseCategoryRepository {
	return &expenseCategoryRepository{}
}

func (r *expenseCategoryRepository) Create(ctx context.Context, tx pgx.Tx, category *domain.ExpenseCategory) error {
	query := `
		INSERT INTO expense_categories (name, description, color, icon, budgeting_group_id, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	category.ExternalID = uuid.New()
	category.CreatedAt = now
	category.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		category.Name,
		category.Description,
		category.Color,
		category.Icon,
		category.BudgetingGroupID,
		category.ExternalID,
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(&category.ID)
}

func (r *expenseCategoryRepository) GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.ExpenseCategory, error) {
	query := `
		SELECT id, external_id, name, description, color, icon, budgeting_group_id, created_at, updated_at, revoked_at
		FROM expense_categories
		WHERE external_id = $1 AND revoked_at IS NULL
	`

	category := &domain.ExpenseCategory{}
	err := tx.QueryRow(ctx, query, externalID).Scan(
		&category.ID,
		&category.ExternalID,
		&category.Name,
		&category.Description,
		&category.Color,
		&category.Icon,
		&category.BudgetingGroupID,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (r *expenseCategoryRepository) GetByGroupID(ctx context.Context, tx pgx.Tx, groupID int64) ([]domain.ExpenseCategory, error) {
	query := `
		SELECT id, external_id, name, description, color, icon, budgeting_group_id, created_at, updated_at, revoked_at
		FROM expense_categories
		WHERE budgeting_group_id = $1 AND revoked_at IS NULL
		ORDER BY name
	`

	rows, err := tx.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.ExpenseCategory
	for rows.Next() {
		var c domain.ExpenseCategory
		if err := rows.Scan(
			&c.ID,
			&c.ExternalID,
			&c.Name,
			&c.Description,
			&c.Color,
			&c.Icon,
			&c.BudgetingGroupID,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.RevokedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, rows.Err()
}

func (r *expenseCategoryRepository) Update(ctx context.Context, tx pgx.Tx, category *domain.ExpenseCategory) error {
	query := `
		UPDATE expense_categories
		SET name = $1, description = $2, color = $3, icon = $4, updated_at = $5
		WHERE id = $6 AND revoked_at IS NULL
	`

	category.UpdatedAt = time.Now()

	result, err := tx.Exec(ctx, query,
		category.Name,
		category.Description,
		category.Color,
		category.Icon,
		category.UpdatedAt,
		category.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *expenseCategoryRepository) Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error {
	query := `
		UPDATE expense_categories
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
