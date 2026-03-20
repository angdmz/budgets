package handler

import (
	"log"
	"net/http"

	"github.com/budgets/core/internal/config"
	"github.com/gin-gonic/gin"
)

// SafeErrorResponse creates an error response that doesn't expose internal details in production
func SafeErrorResponse(c *gin.Context, statusCode int, errorCode string, err error) {
	response := ErrorResponse{
		Error: errorCode,
	}

	// Only include detailed error messages in development
	cfg, exists := c.Get("config")
	if exists {
		if appConfig, ok := cfg.(*config.Config); ok && appConfig.Server.Env.IsDevelopment() {
			response.Message = err.Error()
		}
	}

	// Always log the full error server-side
	if err != nil {
		log.Printf("[ERROR] %s: %v", errorCode, err)
	}

	c.JSON(statusCode, response)
}

// SafeValidationError handles validation errors (these can be shown to users)
func SafeValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "invalid_request",
		Message: err.Error(), // Validation errors are safe to expose
	})
}
