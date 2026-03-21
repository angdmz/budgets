package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/budgets/core/internal/domain"
)

const (
	userContextKey = "user"
)

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: []byte(jwtSecret)}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
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
		user, err := m.validateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "message": err.Error()})
			return
		}

		c.Set(userContextKey, user)
		c.Next()
	}
}

func (m *AuthMiddleware) validateToken(tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	user := &domain.User{
		ExternalProviderID: getLegacyStringClaim(claims, "sub"),
		Email:              getLegacyStringClaim(claims, "email"),
		DisplayName:        getLegacyStringClaim(claims, "name"),
		AuthProvider:       domain.AuthProvider(getLegacyStringClaim(claims, "provider")),
	}

	if user.ExternalProviderID == "" {
		return nil, fmt.Errorf("missing user ID in token")
	}

	return user, nil
}

func (m *AuthMiddleware) GenerateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ExternalProviderID,
		"email":    user.Email,
		"name":     user.DisplayName,
		"provider": string(user.AuthProvider),
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.jwtSecret)
}

func getLegacyStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func GetUserFromContext(c *gin.Context) *domain.User {
	if user, exists := c.Get(userContextKey); exists {
		if u, ok := user.(*domain.User); ok {
			return u
		}
	}
	return nil
}
