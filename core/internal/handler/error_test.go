package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/budgets/core/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSafeErrorResponse_ProductionMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set production config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Env: "production",
		},
	}
	c.Set("config", cfg)

	// Test
	testErr := errors.New("internal database connection failed: timeout")
	SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", testErr)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"internal_error"`)
	assert.NotContains(t, w.Body.String(), "database connection failed")
	assert.NotContains(t, w.Body.String(), "timeout")
	assert.NotContains(t, w.Body.String(), `"message"`)
}

func TestSafeErrorResponse_DevelopmentMode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set development config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Env: "development",
		},
	}
	c.Set("config", cfg)

	// Test
	testErr := errors.New("internal database connection failed: timeout")
	SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", testErr)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"internal_error"`)
	assert.Contains(t, w.Body.String(), `"message":"internal database connection failed: timeout"`)
}

func TestSafeErrorResponse_NoConfigInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Don't set config in context

	// Test
	testErr := errors.New("some internal error")
	SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", testErr)

	// Assert - should default to production behavior (no message)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"internal_error"`)
	assert.NotContains(t, w.Body.String(), "some internal error")
}

func TestSafeErrorResponse_NilError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	cfg := &config.Config{
		Server: config.ServerConfig{
			Env: "production",
		},
	}
	c.Set("config", cfg)

	// Test with nil error
	SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", nil)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"internal_error"`)
}

func TestSafeValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test
	validationErr := errors.New("Key: 'CreateGroupRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag")
	SafeValidationError(c, validationErr)

	// Assert - validation errors should always include message
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"invalid_request"`)
	assert.Contains(t, w.Body.String(), `"message":"Key: 'CreateGroupRequest.Name'`)
}

func TestSafeErrorResponse_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		errorCode  string
	}{
		{"BadRequest", http.StatusBadRequest, "bad_request"},
		{"Unauthorized", http.StatusUnauthorized, "unauthorized"},
		{"Forbidden", http.StatusForbidden, "forbidden"},
		{"NotFound", http.StatusNotFound, "not_found"},
		{"InternalError", http.StatusInternalServerError, "internal_error"},
		{"ServiceUnavailable", http.StatusServiceUnavailable, "service_unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			cfg := &config.Config{
				Server: config.ServerConfig{
					Env: "production",
				},
			}
			c.Set("config", cfg)

			testErr := errors.New("some error")
			SafeErrorResponse(c, tc.statusCode, tc.errorCode, testErr)

			assert.Equal(t, tc.statusCode, w.Code)
			assert.Contains(t, w.Body.String(), `"error":"`+tc.errorCode+`"`)
			assert.NotContains(t, w.Body.String(), "some error")
		})
	}
}
