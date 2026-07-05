package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// UserInfoResponse represents the response from Auth0's /userinfo endpoint.
type UserInfoResponse struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// UserInfoProvider fetches user profile information from the identity provider.
type UserInfoProvider interface {
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfoResponse, error)
}

// Auth0UserInfoProvider calls the Auth0 /userinfo endpoint.
type Auth0UserInfoProvider struct {
	domain string
}

// NewAuth0UserInfoProvider creates a provider that calls https://<domain>/userinfo.
func NewAuth0UserInfoProvider(auth0Domain string) *Auth0UserInfoProvider {
	return &Auth0UserInfoProvider{domain: auth0Domain}
}

func (p *Auth0UserInfoProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfoResponse, error) {
	url := fmt.Sprintf("https://%s/userinfo", p.domain)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call userinfo endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}

	var info UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &info, nil
}
