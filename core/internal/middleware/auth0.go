package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/budgets/core/internal/domain"
)

const (
	auth0UserContextKey = "user"
)

// Auth0Middleware handles Auth0 JWT validation
type Auth0Middleware struct {
	domain    string
	audience  string
	jwksCache *jwksCache
}

// JWKS represents the JSON Web Key Set
type JWKS struct {
	Keys []JSONWebKey `json:"keys"`
}

// JSONWebKey represents a single key in the JWKS
type JSONWebKey struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// jwksCache caches the JWKS to avoid fetching on every request
type jwksCache struct {
	mu         sync.RWMutex
	keys       map[string]*JSONWebKey
	lastUpdate time.Time
	ttl        time.Duration
}

func newJWKSCache(ttl time.Duration) *jwksCache {
	return &jwksCache{
		keys: make(map[string]*JSONWebKey),
		ttl:  ttl,
	}
}

func (c *jwksCache) get(kid string) (*JSONWebKey, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if time.Since(c.lastUpdate) > c.ttl {
		return nil, false
	}
	
	key, ok := c.keys[kid]
	return key, ok
}

func (c *jwksCache) set(keys []JSONWebKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.keys = make(map[string]*JSONWebKey)
	for i := range keys {
		c.keys[keys[i].Kid] = &keys[i]
	}
	c.lastUpdate = time.Now()
}

// NewAuth0Middleware creates a new Auth0 middleware
func NewAuth0Middleware(domain, audience string) *Auth0Middleware {
	return &Auth0Middleware{
		domain:    domain,
		audience:  audience,
		jwksCache: newJWKSCache(1 * time.Hour),
	}
}

// RequireAuth validates Auth0 JWT tokens
func (m *Auth0Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_authorization_header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_authorization_format"})
			return
		}

		tokenString := parts[1]
		user, err := m.validateToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "message": err.Error()})
			return
		}

		c.Set(auth0UserContextKey, user)
		c.Next()
	}
}

func (m *Auth0Middleware) validateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Get the public key from JWKS
		key, err := m.getPublicKey(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	iss, ok := claims["iss"].(string)
	if !ok || iss != fmt.Sprintf("https://%s/", m.domain) {
		return nil, fmt.Errorf("invalid issuer")
	}

	// Validate audience
	if !m.validateAudience(claims) {
		return nil, fmt.Errorf("invalid audience")
	}

	// Validate expiration
	exp, ok := claims["exp"].(float64)
	if !ok || time.Unix(int64(exp), 0).Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	// Extract user information
	user := &domain.User{
		ExternalProviderID: getStringClaim(claims, "sub"),
		Email:              getStringClaim(claims, "email"),
		DisplayName:        getStringClaim(claims, "name"),
		AuthProvider:       domain.AuthProviderGoogle, // Default to Google, can be extended
	}

	if user.ExternalProviderID == "" {
		return nil, fmt.Errorf("missing user ID in token")
	}

	return user, nil
}

func (m *Auth0Middleware) validateAudience(claims jwt.MapClaims) bool {
	aud, ok := claims["aud"]
	if !ok {
		return false
	}

	switch v := aud.(type) {
	case string:
		return v == m.audience
	case []interface{}:
		for _, a := range v {
			if audStr, ok := a.(string); ok && audStr == m.audience {
				return true
			}
		}
	}
	return false
}

func (m *Auth0Middleware) getPublicKey(ctx context.Context, kid string) (interface{}, error) {
	// Check cache first
	if key, ok := m.jwksCache.get(kid); ok {
		return m.parsePublicKey(key)
	}

	// Fetch JWKS from Auth0
	jwks, err := m.fetchJWKS(ctx)
	if err != nil {
		return nil, err
	}

	// Update cache
	m.jwksCache.set(jwks.Keys)

	// Find the key with matching kid
	for i := range jwks.Keys {
		if jwks.Keys[i].Kid == kid {
			return m.parsePublicKey(&jwks.Keys[i])
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

func (m *Auth0Middleware) fetchJWKS(ctx context.Context) (*JWKS, error) {
	url := fmt.Sprintf("https://%s/.well-known/jwks.json", m.domain)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	return &jwks, nil
}

func (m *Auth0Middleware) parsePublicKey(key *JSONWebKey) (interface{}, error) {
	if len(key.X5c) == 0 {
		return nil, fmt.Errorf("no x5c certificate found")
	}

	cert := "-----BEGIN CERTIFICATE-----\n" + key.X5c[0] + "\n-----END CERTIFICATE-----"
	return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

// GetUserFromContext retrieves the authenticated user from the Gin context
func GetAuth0UserFromContext(c *gin.Context) *domain.User {
	if user, exists := c.Get(auth0UserContextKey); exists {
		if u, ok := user.(*domain.User); ok {
			return u
		}
	}
	return nil
}
