package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type budgetingGroupRepository struct{}

func NewBudgetingGroupRepository() BudgetingGroupRepository {
	return &budgetingGroupRepository{}
}

func (r *budgetingGroupRepository) Create(ctx context.Context, tx pgx.Tx, group *domain.BudgetingGroup) error {
	query := `
		INSERT INTO budgeting_groups (name, description, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	group.ExternalID = uuid.New()
	group.CreatedAt = now
	group.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		group.Name,
		group.Description,
		group.ExternalID,
		group.CreatedAt,
		group.UpdatedAt,
	).Scan(&group.ID)
}

func (r *budgetingGroupRepository) GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.BudgetingGroup, error) {
	query := `
		SELECT id, external_id, name, description, created_at, updated_at, revoked_at
		FROM budgeting_groups
		WHERE external_id = $1 AND revoked_at IS NULL
	`

	group := &domain.BudgetingGroup{}
	err := tx.QueryRow(ctx, query, externalID).Scan(
		&group.ID,
		&group.ExternalID,
		&group.Name,
		&group.Description,
		&group.CreatedAt,
		&group.UpdatedAt,
		&group.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (r *budgetingGroupRepository) GetByParticipantUserID(ctx context.Context, tx pgx.Tx, userID string) ([]domain.BudgetingGroup, error) {
	query := `
		SELECT bg.id, bg.external_id, bg.name, bg.description, bg.created_at, bg.updated_at, bg.revoked_at
		FROM budgeting_groups bg
		INNER JOIN participants p ON p.budgeting_group_id = bg.id
		WHERE p.external_user_id = $1 AND bg.revoked_at IS NULL AND p.revoked_at IS NULL
	`

	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []domain.BudgetingGroup
	for rows.Next() {
		var group domain.BudgetingGroup
		if err := rows.Scan(
			&group.ID,
			&group.ExternalID,
			&group.Name,
			&group.Description,
			&group.CreatedAt,
			&group.UpdatedAt,
			&group.RevokedAt,
		); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

func (r *budgetingGroupRepository) Update(ctx context.Context, tx pgx.Tx, group *domain.BudgetingGroup) error {
	query := `
		UPDATE budgeting_groups
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND revoked_at IS NULL
	`

	group.UpdatedAt = time.Now()

	result, err := tx.Exec(ctx, query,
		group.Name,
		group.Description,
		group.UpdatedAt,
		group.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *budgetingGroupRepository) Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error {
	query := `
		UPDATE budgeting_groups
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
