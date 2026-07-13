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

func TestUpdateActualExpense_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &ExpenseHandler{}
	handler.UpdateActualExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestUpdateActualExpense_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	reqBody := UpdateActualExpenseRequest{
		Name:        "Updated Expense",
		Description: "Updated description",
		ExpenseDate: "2025-06-15",
		Amount:      MoneyRequest{Amount: "100.00", Currency: "USD"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("PUT", "/actual-expenses/"+uuid.New().String(), bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &ExpenseHandler{}
	handler.UpdateActualExpense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestUpdateActualExpense_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	user := &domain.User{}
	c.Set("db_user", user)

	c.Request = httptest.NewRequest("PUT", "/actual-expenses/"+uuid.New().String(), bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &ExpenseHandler{}
	handler.UpdateActualExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

func TestCreateExpectedExpense_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "budget_id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	user := &domain.User{}
	c.Set("db_user", user)

	handler := &ExpenseHandler{}
	handler.CreateExpectedExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_budget_id")
}

func TestCreateExpectedExpense_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "budget_id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	reqBody := CreateExpectedExpenseRequest{
		Name:        "New Expected",
		Description: "Description",
		Amount:      MoneyRequest{Amount: "100.00", Currency: "USD"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/budgets/"+uuid.New().String()+"/expected-expenses", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &ExpenseHandler{}
	handler.CreateExpectedExpense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestDeleteActualExpense_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &ExpenseHandler{}
	handler.DeleteActualExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestDeleteActualExpense_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	c.Request = httptest.NewRequest("DELETE", "/actual-expenses/"+uuid.New().String(), nil)

	handler := &ExpenseHandler{}
	handler.DeleteActualExpense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestUpdateExpectedExpense_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &ExpenseHandler{}
	handler.UpdateExpectedExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestUpdateExpectedExpense_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	reqBody := UpdateExpectedExpenseRequest{
		Name:        "Updated Expected",
		Description: "Updated description",
		Amount:      MoneyRequest{Amount: "100.00", Currency: "USD"},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("PUT", "/expected-expenses/"+uuid.New().String(), bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := &ExpenseHandler{}
	handler.UpdateExpectedExpense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestDeleteExpectedExpense_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	handler := &ExpenseHandler{}
	handler.DeleteExpectedExpense(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_id")
}

func TestDeleteExpectedExpense_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	cfg := &config.Config{Server: config.ServerConfig{Env: "test"}}
	c.Set("config", cfg)

	c.Request = httptest.NewRequest("DELETE", "/expected-expenses/"+uuid.New().String(), nil)

	handler := &ExpenseHandler{}
	handler.DeleteExpectedExpense(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}
