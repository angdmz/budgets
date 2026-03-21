package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/budgets/core/internal/domain"
)

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) GetOrCreateByProvider(ctx context.Context, tx pgx.Tx, providerID string, provider domain.AuthProvider, email, displayName, avatarURL string) (*domain.User, error) {
	// Try to find existing user by provider
	user, err := r.getByProviderID(ctx, tx, provider, providerID)
	if err == nil {
		// Update profile info if changed
		if user.Email != email || user.DisplayName != displayName || user.AvatarURL != avatarURL {
			user.Email = email
			user.DisplayName = displayName
			user.AvatarURL = avatarURL
			user.UpdatedAt = time.Now()
			if updateErr := r.update(ctx, tx, user); updateErr != nil {
				return nil, updateErr
			}
		}
		return user, nil
	}

	// Create new user
	user = &domain.User{
		ExternalProviderID: providerID,
		AuthProvider:       provider,
		Email:              email,
		DisplayName:        displayName,
		AvatarURL:          avatarURL,
	}

	now := time.Now()
	user.ExternalID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (external_provider_id, auth_provider, email, display_name, avatar_url, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err = tx.QueryRow(ctx, query,
		user.ExternalProviderID,
		string(user.AuthProvider),
		user.Email,
		user.DisplayName,
		user.AvatarURL,
		user.ExternalID,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, tx pgx.Tx, id int64) (*domain.User, error) {
	query := `
		SELECT id, external_id, external_provider_id, auth_provider, email, display_name, avatar_url, created_at, updated_at, revoked_at
		FROM users
		WHERE id = $1 AND revoked_at IS NULL
	`

	return r.scanUser(tx.QueryRow(ctx, query, id))
}

func (r *userRepository) GetByExternalID(ctx context.Context, tx pgx.Tx, externalID uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, external_id, external_provider_id, auth_provider, email, display_name, avatar_url, created_at, updated_at, revoked_at
		FROM users
		WHERE external_id = $1 AND revoked_at IS NULL
	`

	return r.scanUser(tx.QueryRow(ctx, query, externalID))
}

func (r *userRepository) getByProviderID(ctx context.Context, tx pgx.Tx, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	query := `
		SELECT id, external_id, external_provider_id, auth_provider, email, display_name, avatar_url, created_at, updated_at, revoked_at
		FROM users
		WHERE auth_provider = $1 AND external_provider_id = $2 AND revoked_at IS NULL
	`

	return r.scanUser(tx.QueryRow(ctx, query, string(provider), providerID))
}

func (r *userRepository) update(ctx context.Context, tx pgx.Tx, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, display_name = $2, avatar_url = $3, updated_at = $4
		WHERE id = $5 AND revoked_at IS NULL
	`

	_, err := tx.Exec(ctx, query,
		user.Email,
		user.DisplayName,
		user.AvatarURL,
		user.UpdatedAt,
		user.ID,
	)
	return err
}

func (r *userRepository) scanUser(row pgx.Row) (*domain.User, error) {
	user := &domain.User{}
	var authProvider string
	err := row.Scan(
		&user.ID,
		&user.ExternalID,
		&user.ExternalProviderID,
		&authProvider,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.RevokedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	user.AuthProvider = domain.AuthProvider(authProvider)
	return user, nil
}
