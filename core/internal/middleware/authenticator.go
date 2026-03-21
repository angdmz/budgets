package middleware

import "github.com/gin-gonic/gin"

// Authenticator is the interface that both Auth0Middleware and AuthMiddleware implement.
// It allows the server to be configured with either one, enabling integration tests
// to use the simpler JWT-based AuthMiddleware.
type Authenticator interface {
	RequireAuth() gin.HandlerFunc
}
