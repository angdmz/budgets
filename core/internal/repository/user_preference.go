package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type userPreferenceRepository struct{}

func NewUserPreferenceRepository() UserPreferenceRepository {
	return &userPreferenceRepository{}
}

func (r *userPreferenceRepository) GetByUserID(ctx context.Context, tx pgx.Tx, userID int64) (*domain.UserPreference, error) {
	query := `
		SELECT id, external_id, user_id, theme, language, display_currency, created_at, updated_at, revoked_at
		FROM user_preferences
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	pref := &domain.UserPreference{}
	err := tx.QueryRow(ctx, query, userID).Scan(
		&pref.ID, &pref.ExternalID, &pref.UserID,
		&pref.Theme, &pref.Language, &pref.DisplayCurrency,
		&pref.CreatedAt, &pref.UpdatedAt, &pref.RevokedAt,
	)
	if err != nil {
		return nil, err
	}
	return pref, nil
}

func (r *userPreferenceRepository) Upsert(ctx context.Context, tx pgx.Tx, pref *domain.UserPreference) error {
	now := time.Now()

	query := `
		INSERT INTO user_preferences (external_id, user_id, theme, language, display_currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) WHERE revoked_at IS NULL
		DO UPDATE SET theme = EXCLUDED.theme, language = EXCLUDED.language, display_currency = EXCLUDED.display_currency, updated_at = EXCLUDED.updated_at
		RETURNING id, external_id, created_at, updated_at
	`

	externalID := uuid.New()
	return tx.QueryRow(ctx, query,
		externalID, pref.UserID, pref.Theme, pref.Language, pref.DisplayCurrency,
		now, now,
	).Scan(&pref.ID, &pref.ExternalID, &pref.CreatedAt, &pref.UpdatedAt)
}
