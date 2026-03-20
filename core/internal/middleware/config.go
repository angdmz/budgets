package middleware

import (
	"github.com/budgets/core/internal/config"
	"github.com/gin-gonic/gin"
)

// InjectConfig adds the application config to the Gin context
func InjectConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	}
}
