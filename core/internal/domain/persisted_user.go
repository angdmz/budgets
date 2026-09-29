package domain

import (
	"context"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/budgets/core/internal/representation"
)

// PersistedUser is a user row loaded from (or upserted into) the database.
// Its internal id stays unexported; it only ever crosses into SQL through
// driver.Valuer, which binds it as a user_id argument.
type PersistedUser struct {
	id                 int64
	externalID         uuid.UUID
	externalProviderID string
	authProvider       AuthProvider
	email              string
	displayName        string
	avatarURL          string
	createdAt          time.Time
	updatedAt          time.Time
}

// Value implements driver.Valuer so a *PersistedUser can be passed directly
// as a SQL argument wherever a user_id column is expected, without exposing
// the internal id to callers.
func (u *PersistedUser) Value() (driver.Value, error) {
	return u.id, nil
}

// PersistedUserForProvider upserts a user row for the given identity-provider
// credentials in a single statement. Profile fields are refreshed only when
// the incoming value is non-empty — the provider never wipes stored data.
func PersistedUserForProvider(
	ctx context.Context,
	providerID string,
	provider AuthProvider,
	email, displayName, avatarURL string,
	p Persister,
) (*PersistedUser, error) {
	if providerID == "" {
		return nil, fmt.Errorf("%w: external provider id cannot be empty", ErrValidation)
	}

	var emailArg *string
	if email != "" {
		emailArg = &email
	}

	var u PersistedUser
	err := p.QueryRow(
		ctx,
		[]any{&u.id, &u.externalID, &u.externalProviderID, &u.authProvider, &u.email, &u.displayName, &u.avatarURL, &u.createdAt, &u.updatedAt},
		`INSERT INTO users (external_provider_id, auth_provider, email, display_name, avatar_url)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (auth_provider, external_provider_id) DO UPDATE
			SET email = COALESCE(NULLIF(EXCLUDED.email, ''), users.email),
				display_name = COALESCE(NULLIF(EXCLUDED.display_name, ''), users.display_name),
				avatar_url = COALESCE(NULLIF(EXCLUDED.avatar_url, ''), users.avatar_url),
				updated_at = CURRENT_TIMESTAMP
		 RETURNING id, external_id, external_provider_id, auth_provider, email, display_name, avatar_url, created_at, updated_at`,
		providerID, string(provider), emailArg, displayName, avatarURL,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Render returns the final wire representation of this user.
func (u *PersistedUser) Render() Rendered[representation.User] {
	return render(representation.User{
		ID:          u.externalID,
		Provider:    string(u.authProvider),
		Email:       u.email,
		DisplayName: u.displayName,
		AvatarURL:   u.avatarURL,
		CreatedAt:   u.createdAt,
		UpdatedAt:   u.updatedAt,
	})
}
