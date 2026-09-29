package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
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
type UserResolverFunc func(ctx context.Context, providerID string, provider domain.AuthProvider, email, displayName, avatarURL string, p domain.Persister) (*domain.PersistedUser, error)

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

		provider := domain.AuthProvider(strings.ToUpper(string(authUser.AuthProvider)))
		if provider == "" {
			provider = domain.AuthProviderGoogle
		}

		// Resolve (get-or-create) the DB user within a transaction
		var dbUser *domain.PersistedUser
		err := database.WithPersister(c.Request.Context(), ur.pool, func(ctx context.Context, p *database.PgxPersister) error {
			var resolveErr error
			dbUser, resolveErr = ur.resolveFunc(ctx, authUser.ExternalProviderID, provider, authUser.Email, authUser.DisplayName, authUser.AvatarURL, p)
			return resolveErr
		})
		if err != nil {
			log.Printf("[ERROR] user_resolver: user resolution failed: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user_resolution_failed"})
			return
		}

		c.Set(dbUserContextKey, dbUser)
		c.Next()
	}
}

// GetDBUserFromContext retrieves the resolved DB user from the Gin context.
func GetDBUserFromContext(c *gin.Context) *domain.PersistedUser {
	if user, exists := c.Get(dbUserContextKey); exists {
		if u, ok := user.(*domain.PersistedUser); ok {
			return u
		}
	}
	return nil
}
