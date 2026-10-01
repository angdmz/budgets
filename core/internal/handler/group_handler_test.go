package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/budgets/core/internal/config"
	"github.com/budgets/core/internal/domain"
)

// Regression: CreateGroup must return 400 (not 500) when the request fails
// validation (handoff Phase 0, item 5).
func TestCreateGroup_EmptyName_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)
	c.Set("db_user", &domain.PersistedUser{})

	bodyBytes, _ := json.Marshal(CreateGroupRequest{Description: "no name"})
	c.Request = httptest.NewRequest("POST", "/groups", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &GroupHandler{}
	handler.CreateGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGroup_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	bodyBytes, _ := json.Marshal(CreateGroupRequest{Name: "Family"})
	c.Request = httptest.NewRequest("POST", "/groups", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &GroupHandler{}
	handler.CreateGroup(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestDeleteGroup_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)
	c.Set("db_user", &domain.PersistedUser{})

	handler := &GroupHandler{}
	handler.DeleteGroup(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestDeleteGroup_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	c.Request = httptest.NewRequest("DELETE", "/groups/"+uuid.New().String(), nil)

	handler := &GroupHandler{}
	handler.DeleteGroup(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}
