package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
