package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/budgets/core/internal/domain"
)

func TestAuthMiddleware_GenerateAndValidateToken(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret-key-32-chars-long!!")

	user := &domain.User{
		ExternalProviderID: "user-123",
		Email:              "test@example.com",
		DisplayName:        "Test User",
		AuthProvider:       "google",
	}

	token, err := middleware.GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	validatedUser, err := middleware.validateToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ExternalProviderID, validatedUser.ExternalProviderID)
	assert.Equal(t, user.Email, validatedUser.Email)
	assert.Equal(t, user.DisplayName, validatedUser.DisplayName)
	assert.Equal(t, user.AuthProvider, validatedUser.AuthProvider)
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware := NewAuthMiddleware("test-secret-key-32-chars-long!!")

	user := &domain.User{
		ExternalProviderID: "user-123",
		Email:              "test@example.com",
		DisplayName:        "Test User",
		AuthProvider:       "google",
	}

	token, err := middleware.GenerateToken(user)
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequireAuth())
		router.GET("/test", func(c *gin.Context) {
			u := GetUserFromContext(c)
			c.JSON(200, gin.H{"user_id": u.ExternalProviderID})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequireAuth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequireAuth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"ok": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestSSOLoginStrategy_ExtractUser(t *testing.T) {
	t.Run("extracts_all_fields", func(t *testing.T) {
		strategy := &SSOLoginStrategy{provider: domain.AuthProviderGoogle}
		claims := jwt.MapClaims{
			"sub":     "google-oauth2|12345",
			"email":   "user@gmail.com",
			"name":    "Google User",
			"picture": "https://example.com/pic.jpg",
		}
		user, err := strategy.ExtractUser(claims)
		require.NoError(t, err)
		assert.Equal(t, "google-oauth2|12345", user.ExternalProviderID)
		assert.Equal(t, "user@gmail.com", user.Email)
		assert.Equal(t, "Google User", user.DisplayName)
		assert.Equal(t, "https://example.com/pic.jpg", user.AvatarURL)
		assert.Equal(t, domain.AuthProviderGoogle, user.AuthProvider)
	})

	t.Run("allows_missing_email", func(t *testing.T) {
		strategy := &SSOLoginStrategy{provider: domain.AuthProviderGoogle}
		claims := jwt.MapClaims{
			"sub": "google-oauth2|12345",
		}
		user, err := strategy.ExtractUser(claims)
		require.NoError(t, err)
		assert.Equal(t, "", user.Email)
		assert.Equal(t, "", user.DisplayName)
	})

	t.Run("fails_without_sub", func(t *testing.T) {
		strategy := &SSOLoginStrategy{provider: domain.AuthProviderGoogle}
		claims := jwt.MapClaims{
			"email": "user@gmail.com",
		}
		_, err := strategy.ExtractUser(claims)
		require.Error(t, err)
	})
}

func TestUsernamePasswordLoginStrategy_ExtractUser(t *testing.T) {
	t.Run("uses_email_as_display_name_when_name_absent", func(t *testing.T) {
		strategy := &UsernamePasswordLoginStrategy{}
		claims := jwt.MapClaims{
			"sub":   "auth0|abc123",
			"email": "user@example.com",
		}
		user, err := strategy.ExtractUser(claims)
		require.NoError(t, err)
		assert.Equal(t, "auth0|abc123", user.ExternalProviderID)
		assert.Equal(t, "user@example.com", user.Email)
		assert.Equal(t, "user@example.com", user.DisplayName)
		assert.Equal(t, domain.AuthProviderLocal, user.AuthProvider)
	})

	t.Run("uses_name_claim_when_present", func(t *testing.T) {
		strategy := &UsernamePasswordLoginStrategy{}
		claims := jwt.MapClaims{
			"sub":   "auth0|abc123",
			"email": "user@example.com",
			"name":  "John Doe",
		}
		user, err := strategy.ExtractUser(claims)
		require.NoError(t, err)
		assert.Equal(t, "John Doe", user.DisplayName)
	})

	t.Run("succeeds_without_email", func(t *testing.T) {
		strategy := &UsernamePasswordLoginStrategy{}
		claims := jwt.MapClaims{
			"sub": "auth0|abc123",
		}
		user, err := strategy.ExtractUser(claims)
		require.NoError(t, err)
		assert.Equal(t, "auth0|abc123", user.ExternalProviderID)
		assert.Equal(t, "", user.Email)
		assert.Equal(t, "", user.DisplayName)
	})

	t.Run("fails_without_sub", func(t *testing.T) {
		strategy := &UsernamePasswordLoginStrategy{}
		claims := jwt.MapClaims{
			"email": "user@example.com",
		}
		_, err := strategy.ExtractUser(claims)
		require.Error(t, err)
	})
}

func TestDetectAuthProvider(t *testing.T) {
	tests := []struct {
		sub      string
		expected domain.AuthProvider
	}{
		{"google-oauth2|12345", domain.AuthProviderGoogle},
		{"github|67890", domain.AuthProviderGitHub},
		{"auth0|abc123", domain.AuthProviderLocal},
		{"unknown|xyz", domain.AuthProviderLocal},
		{"", domain.AuthProviderLocal},
	}

	for _, tc := range tests {
		t.Run(tc.sub, func(t *testing.T) {
			result := detectAuthProvider(tc.sub)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestUserResolver_Auth0PasswordUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Simulate what happens when a user logs in via Auth0 username/password:
	// - sub = "auth0|abc123" (LOCAL provider)
	// - email is the username (required); display name defaults to email when absent.

	authMiddleware := NewAuthMiddleware("test-secret-key-32-chars-long!!")

	// Mock resolve function that tracks calls
	var resolveCalls []struct {
		providerID  string
		provider    domain.AuthProvider
		email       string
		displayName string
	}
	mockResolveFunc := func(ctx context.Context, tx pgx.Tx, providerID string, provider domain.AuthProvider, email, displayName, avatarURL string) (*domain.User, error) {
		resolveCalls = append(resolveCalls, struct {
			providerID  string
			provider    domain.AuthProvider
			email       string
			displayName string
		}{providerID, provider, email, displayName})
		return &domain.User{
			ExternalProviderID: providerID,
			AuthProvider:       provider,
			Email:              email,
			DisplayName:        displayName,
		}, nil
	}

	t.Run("auth0_password_user_resolves_as_LOCAL", func(t *testing.T) {
		resolveCalls = nil

		// Create a token simulating an auth0| user (password login).
		// Email is the username; display name defaults to email when absent.
		testUser := &domain.User{
			ExternalProviderID: "auth0|6a2eb82f8451dc995b556d08",
			Email:              "user@example.com",
			DisplayName:        "user@example.com",
			AuthProvider:       domain.AuthProviderLocal,
		}
		token, err := authMiddleware.GenerateToken(testUser)
		require.NoError(t, err)

		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		// Use a mock user resolver inline (bypasses DB)
		router.Use(func(c *gin.Context) {
			authUser := GetUserFromContext(c)
			if authUser == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no_user"})
				return
			}
			// Simulate the resolver logic
			provider := authUser.AuthProvider
			_, _ = mockResolveFunc(c.Request.Context(), nil, authUser.ExternalProviderID, provider, authUser.Email, authUser.DisplayName, "")
			c.Next()
		})
		router.GET("/api/v1/groups", func(c *gin.Context) {
			c.JSON(200, gin.H{"groups": []string{}})
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify the resolver was called with LOCAL provider and auth0| ID
		require.Len(t, resolveCalls, 1)
		assert.Equal(t, "auth0|6a2eb82f8451dc995b556d08", resolveCalls[0].providerID)
		assert.Equal(t, domain.AuthProviderLocal, resolveCalls[0].provider)
	})

	t.Run("google_oauth2_user_resolves_as_GOOGLE", func(t *testing.T) {
		resolveCalls = nil

		testUser := &domain.User{
			ExternalProviderID: "google-oauth2|114928371234",
			Email:              "user@gmail.com",
			DisplayName:        "Google User",
			AuthProvider:       domain.AuthProviderGoogle,
		}
		token, err := authMiddleware.GenerateToken(testUser)
		require.NoError(t, err)

		router := gin.New()
		router.Use(authMiddleware.RequireAuth())
		router.Use(func(c *gin.Context) {
			authUser := GetUserFromContext(c)
			if authUser == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no_user"})
				return
			}
			provider := authUser.AuthProvider
			_, _ = mockResolveFunc(c.Request.Context(), nil, authUser.ExternalProviderID, provider, authUser.Email, authUser.DisplayName, "")
			c.Next()
		})
		router.GET("/api/v1/groups", func(c *gin.Context) {
			c.JSON(200, gin.H{"groups": []string{}})
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		require.Len(t, resolveCalls, 1)
		assert.Equal(t, "google-oauth2|114928371234", resolveCalls[0].providerID)
		assert.Equal(t, domain.AuthProviderGoogle, resolveCalls[0].provider)
		assert.Equal(t, "user@gmail.com", resolveCalls[0].email)
	})
}
