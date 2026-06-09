package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type UserRepository interface {
	GetOrCreateByProvider(ctx context.Context, tx pgx.Tx, providerID string, provider domain.AuthProvider, email, displayName, avatarURL string) (*domain.User, error)
	GetByID(ctx context.Context, tx pgx.Tx, id int64) (*domain.User, error)
	GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.User, error)
}

type UserPreferenceRepository interface {
	GetByUserID(ctx context.Context, tx pgx.Tx, userID int64) (*domain.UserPreference, error)
	Upsert(ctx context.Context, tx pgx.Tx, pref *domain.UserPreference) error
}
