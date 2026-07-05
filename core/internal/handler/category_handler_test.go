package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/budgets/core/internal/config"
	"github.com/budgets/core/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCategory_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &CategoryHandler{}
	handler.UpdateCategory(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestUpdateCategory_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	reqBody := UpdateCategoryRequest{
		Name:        "Updated Category",
		Description: "Updated description",
		Color:       "#ff0000",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("PUT", "/categories/"+uuid.New().String(), bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &CategoryHandler{}
	handler.UpdateCategory(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestUpdateCategory_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	user := &domain.User{}
	c.Set("db_user", user)

	c.Request = httptest.NewRequest("PUT", "/categories/"+uuid.New().String(), bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &CategoryHandler{}
	handler.UpdateCategory(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

func TestDeleteCategory_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &CategoryHandler{}
	handler.DeleteCategory(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestDeleteCategory_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	c.Request = httptest.NewRequest("DELETE", "/categories/"+uuid.New().String(), nil)

	handler := &CategoryHandler{}
	handler.DeleteCategory(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}
