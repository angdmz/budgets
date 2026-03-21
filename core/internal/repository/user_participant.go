package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type userParticipantRepository struct{}

func NewUserParticipantRepository() UserParticipantRepository {
	return &userParticipantRepository{}
}

func (r *userParticipantRepository) Create(ctx context.Context, tx pgx.Tx, up *domain.UserParticipant) error {
	query := `
		INSERT INTO user_participants (user_id, participant_id, role, is_primary, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	up.ExternalID = uuid.New()
	up.CreatedAt = now
	up.UpdatedAt = now

	if up.Role == "" {
		up.Role = "owner"
	}

	isPrimary := 0
	if up.IsPrimary {
		isPrimary = 1
	}

	return tx.QueryRow(ctx, query,
		up.UserID,
		up.ParticipantID,
		up.Role,
		isPrimary,
		up.ExternalID,
		up.CreatedAt,
		up.UpdatedAt,
	).Scan(&up.ID)
}

func (r *userParticipantRepository) GetByUserIDAndGroupID(ctx context.Context, tx pgx.Tx, userID int64, groupID int64) (*domain.UserParticipant, error) {
	query := `
		SELECT up.id, up.external_id, up.user_id, up.participant_id, up.role, up.is_primary, up.created_at, up.updated_at, up.revoked_at
		FROM user_participants up
		INNER JOIN participants p ON p.id = up.participant_id
		WHERE up.user_id = $1 AND p.budgeting_group_id = $2 AND up.revoked_at IS NULL AND p.revoked_at IS NULL
	`

	up := &domain.UserParticipant{}
	var isPrimary int
	err := tx.QueryRow(ctx, query, userID, groupID).Scan(
		&up.ID,
		&up.ExternalID,
		&up.UserID,
		&up.ParticipantID,
		&up.Role,
		&isPrimary,
		&up.CreatedAt,
		&up.UpdatedAt,
		&up.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	up.IsPrimary = isPrimary == 1
	return up, nil
}

func (r *userParticipantRepository) GetByUserIDAndParticipantID(ctx context.Context, tx pgx.Tx, userID, participantID int64) (*domain.UserParticipant, error) {
	query := `
		SELECT id, external_id, user_id, participant_id, role, is_primary, created_at, updated_at, revoked_at
		FROM user_participants
		WHERE user_id = $1 AND participant_id = $2 AND revoked_at IS NULL
	`

	up := &domain.UserParticipant{}
	var isPrimary int
	err := tx.QueryRow(ctx, query, userID, participantID).Scan(
		&up.ID,
		&up.ExternalID,
		&up.UserID,
		&up.ParticipantID,
		&up.Role,
		&isPrimary,
		&up.CreatedAt,
		&up.UpdatedAt,
		&up.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	up.IsPrimary = isPrimary == 1
	return up, nil
}
