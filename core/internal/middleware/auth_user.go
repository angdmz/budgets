package middleware

import (
	"github.com/budgets/core/internal/domain"
)

// AuthUser is the thin middleware DTO carrying identity-provider claims
// between token validation and user resolution. It is not a domain object:
// it lives only inside gin contexts on the auth path.
type AuthUser struct {
	ExternalProviderID string
	Email              string
	DisplayName        string
	AvatarURL          string
	AuthProvider       domain.AuthProvider
}
