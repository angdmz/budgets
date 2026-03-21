package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/domain"
)

const (
	dbUserContextKey = "db_user"
)

// UserResolver is an interface for resolving/upserting a DB user from auth claims.
type UserResolver interface {
	ResolveUser() gin.HandlerFunc
}

// UserResolverFunc is a function type that can get-or-create a user within a transaction.
type UserResolverFunc func(ctx context.Context, tx pgx.Tx, providerID string, provider domain.AuthProvider, email, displayName, avatarURL string) (*domain.User, error)

type userResolver struct {
	pool       *pgxpool.Pool
	resolveFunc UserResolverFunc
}

// NewUserResolver creates a middleware that upserts a DB user from the authenticated token claims.
func NewUserResolver(pool *pgxpool.Pool, resolveFunc UserResolverFunc) UserResolver {
	return &userResolver{
		pool:       pool,
		resolveFunc: resolveFunc,
	}
}

func (ur *userResolver) ResolveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the auth user set by the auth middleware (both Auth0 and test use same context key)
		authUser := GetAuth0UserFromContext(c)
		if authUser == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no_authenticated_user"})
			return
		}

		// Resolve (get-or-create) the DB user within a transaction
		tx, err := ur.pool.Begin(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "db_error"})
			return
		}
		defer tx.Rollback(c.Request.Context())

		provider := domain.AuthProvider(strings.ToUpper(string(authUser.AuthProvider)))
		if provider == "" {
			provider = domain.AuthProviderGoogle
		}

		dbUser, err := ur.resolveFunc(c.Request.Context(), tx, authUser.ExternalProviderID, provider, authUser.Email, authUser.DisplayName, authUser.AvatarURL)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user_resolution_failed"})
			return
		}

		if err := tx.Commit(c.Request.Context()); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "db_commit_failed"})
			return
		}

		c.Set(dbUserContextKey, dbUser)
		c.Next()
	}
}

// GetDBUserFromContext retrieves the resolved DB user from the Gin context.
func GetDBUserFromContext(c *gin.Context) *domain.User {
	if user, exists := c.Get(dbUserContextKey); exists {
		if u, ok := user.(*domain.User); ok {
			return u
		}
	}
	return nil
}
