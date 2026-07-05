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

func TestUpdateBudget_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &BudgetHandler{}
	handler.UpdateBudget(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestUpdateBudget_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	reqBody := UpdateBudgetRequest{
		Name:        "Updated Budget",
		Description: "Updated description",
		StartDate:   "2025-01-01",
		EndDate:     "2025-12-31",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("PUT", "/budgets/"+uuid.New().String(), bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &BudgetHandler{}
	handler.UpdateBudget(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestUpdateBudget_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	user := &domain.User{}
	c.Set("db_user", user)

	c.Request = httptest.NewRequest("PUT", "/budgets/"+uuid.New().String(), bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &BudgetHandler{}
	handler.UpdateBudget(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

func TestUpdateBudget_InvalidDateFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	user := &domain.User{}
	c.Set("db_user", user)

	reqBody := UpdateBudgetRequest{
		Name:        "Updated Budget",
		Description: "Updated description",
		StartDate:   "invalid-date",
		EndDate:     "2025-12-31",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("PUT", "/budgets/"+uuid.New().String(), bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &BudgetHandler{}
	handler.UpdateBudget(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_start_date")
}

func TestDeleteBudget_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &BudgetHandler{}
	handler.DeleteBudget(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestDeleteBudget_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	c.Request = httptest.NewRequest("DELETE", "/budgets/"+uuid.New().String(), nil)

	handler := &BudgetHandler{}
	handler.DeleteBudget(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}
