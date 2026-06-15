package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/budgets/core/internal/domain"
)

// LoginStrategy extracts a domain.User from validated JWT claims
// based on the login type detected from the sub claim prefix.
type LoginStrategy interface {
	ExtractUser(claims jwt.MapClaims) (*domain.User, error)
}

// SSOLoginStrategy handles google-oauth2| and github| logins.
// Expects email to be present in claims.
type SSOLoginStrategy struct {
	provider domain.AuthProvider
}

func (s *SSOLoginStrategy) ExtractUser(claims jwt.MapClaims) (*domain.User, error) {
	sub := getStringClaim(claims, "sub")
	if sub == "" {
		return nil, fmt.Errorf("missing sub claim")
	}
	email := getStringClaim(claims, "email")
	if email == "" {
		return nil, fmt.Errorf("missing email claim for SSO login")
	}
	return &domain.User{
		ExternalProviderID: sub,
		Email:              email,
		DisplayName:        getStringClaim(claims, "name"),
		AvatarURL:          getStringClaim(claims, "picture"),
		AuthProvider:       s.provider,
	}, nil
}

// UsernamePasswordLoginStrategy handles auth0| (Database Connection) logins.
// The username is the email; uses email as display name when name is absent.
type UsernamePasswordLoginStrategy struct{}

func (s *UsernamePasswordLoginStrategy) ExtractUser(claims jwt.MapClaims) (*domain.User, error) {
	sub := getStringClaim(claims, "sub")
	if sub == "" {
		return nil, fmt.Errorf("missing sub claim")
	}
	email := getStringClaim(claims, "email")
	displayName := getStringClaim(claims, "name")
	if displayName == "" {
		displayName = email
	}
	return &domain.User{
		ExternalProviderID: sub,
		Email:              email,
		DisplayName:        displayName,
		AuthProvider:       domain.AuthProviderLocal,
	}, nil
}
