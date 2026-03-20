package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type participantRepository struct{}

func NewParticipantRepository() ParticipantRepository {
	return &participantRepository{}
}

func (r *participantRepository) Create(ctx context.Context, tx pgx.Tx, participant *domain.Participant) error {
	query := `
		INSERT INTO participants (external_user_id, email, display_name, budgeting_group_id, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	participant.ExternalID = uuid.New()
	participant.CreatedAt = now
	participant.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		participant.ExternalUserID,
		participant.Email,
		participant.DisplayName,
		participant.BudgetingGroupID,
		participant.ExternalID,
		participant.CreatedAt,
		participant.UpdatedAt,
	).Scan(&participant.ID)
}

func (r *participantRepository) GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.Participant, error) {
	query := `
		SELECT id, external_id, external_user_id, email, display_name, budgeting_group_id, created_at, updated_at, revoked_at
		FROM participants
		WHERE external_id = $1 AND revoked_at IS NULL
	`

	participant := &domain.Participant{}
	err := tx.QueryRow(ctx, query, externalID).Scan(
		&participant.ID,
		&participant.ExternalID,
		&participant.ExternalUserID,
		&participant.Email,
		&participant.DisplayName,
		&participant.BudgetingGroupID,
		&participant.CreatedAt,
		&participant.UpdatedAt,
		&participant.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return participant, nil
}

func (r *participantRepository) GetByUserIDAndGroupID(ctx context.Context, tx pgx.Tx, userID string, groupID int64) (*domain.Participant, error) {
	query := `
		SELECT id, external_id, external_user_id, email, display_name, budgeting_group_id, created_at, updated_at, revoked_at
		FROM participants
		WHERE external_user_id = $1 AND budgeting_group_id = $2 AND revoked_at IS NULL
	`

	participant := &domain.Participant{}
	err := tx.QueryRow(ctx, query, userID, groupID).Scan(
		&participant.ID,
		&participant.ExternalID,
		&participant.ExternalUserID,
		&participant.Email,
		&participant.DisplayName,
		&participant.BudgetingGroupID,
		&participant.CreatedAt,
		&participant.UpdatedAt,
		&participant.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return participant, nil
}

func (r *participantRepository) GetByGroupID(ctx context.Context, tx pgx.Tx, groupID int64) ([]domain.Participant, error) {
	query := `
		SELECT id, external_id, external_user_id, email, display_name, budgeting_group_id, created_at, updated_at, revoked_at
		FROM participants
		WHERE budgeting_group_id = $1 AND revoked_at IS NULL
	`

	rows, err := tx.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []domain.Participant
	for rows.Next() {
		var p domain.Participant
		if err := rows.Scan(
			&p.ID,
			&p.ExternalID,
			&p.ExternalUserID,
			&p.Email,
			&p.DisplayName,
			&p.BudgetingGroupID,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.RevokedAt,
		); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}

	return participants, rows.Err()
}

func (r *participantRepository) Update(ctx context.Context, tx pgx.Tx, participant *domain.Participant) error {
	query := `
		UPDATE participants
		SET email = $1, display_name = $2, updated_at = $3
		WHERE id = $4 AND revoked_at IS NULL
	`

	participant.UpdatedAt = time.Now()

	result, err := tx.Exec(ctx, query,
		participant.Email,
		participant.DisplayName,
		participant.UpdatedAt,
		participant.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *participantRepository) Revoke(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) error {
	query := `
		UPDATE participants
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
