package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpenseAPI(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	// Setup: group → category → budget
	var groupID, budgetID, categoryID string
	t.Run("Setup", func(t *testing.T) {
		resp := ts.Post("/api/v1/groups", map[string]interface{}{
			"name": "Expense Test Group",
		})
		require.Equal(t, http.StatusCreated, resp.Code)
		var g map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &g))
		groupID = g["id"].(string)

		resp = ts.Post("/api/v1/groups/"+groupID+"/categories", map[string]interface{}{
			"name":  "Test Category",
			"color": "#0ea5e9",
		})
		require.Equal(t, http.StatusCreated, resp.Code)
		var c map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &c))
		categoryID = c["id"].(string)

		now := time.Now()
		resp = ts.Post("/api/v1/groups/"+groupID+"/budgets", map[string]interface{}{
			"name":       "Test Budget",
			"start_date": now.Format("2006-01-02"),
			"end_date":   now.AddDate(0, 1, 0).Format("2006-01-02"),
		})
		require.Equal(t, http.StatusCreated, resp.Code)
		var b map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &b))
		budgetID = b["id"].(string)
	})

	// --- Expected Expenses ---
	var expectedID string

	t.Run("CreateExpectedExpense", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Rent",
			"description": "Monthly rent",
			"amount": map[string]interface{}{
				"amount":   "1500.00",
				"currency": "USD",
			},
			"category_id": categoryID,
		}
		resp := ts.Post("/api/v1/budgets/"+budgetID+"/expected-expenses", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Equal(t, "Rent", result["name"])
		amt := result["amount"].(map[string]interface{})
		assert.Equal(t, "1500", amt["amount"])
		assert.Equal(t, "USD", amt["currency"])
		assert.Equal(t, categoryID, result["category_id"])
		expectedID = result["id"].(string)
	})

	t.Run("GetExpectedExpense", func(t *testing.T) {
		resp := ts.Get("/api/v1/expected-expenses/" + expectedID)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Equal(t, expectedID, result["id"])
		assert.Equal(t, "Rent", result["name"])
		assert.Equal(t, categoryID, result["category_id"])
	})

	t.Run("ListExpectedExpenses", func(t *testing.T) {
		resp := ts.Get("/api/v1/budgets/" + budgetID + "/expected-expenses")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.GreaterOrEqual(t, len(result), 1)
	})

	t.Run("UpdateExpectedExpense", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Rent + Utilities",
			"description": "Updated",
			"amount": map[string]interface{}{
				"amount":   "1800.50",
				"currency": "USD",
			},
			"category_id": categoryID,
		}
		resp := ts.Put("/api/v1/expected-expenses/"+expectedID, body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Equal(t, "Rent + Utilities", result["name"])
		amt := result["amount"].(map[string]interface{})
		assert.Equal(t, "1800.5", amt["amount"])
		assert.Equal(t, categoryID, result["category_id"])
	})

	t.Run("CreateExpectedExpenseWithoutCategory", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "No Category Expense",
			"description": "Test without category",
			"amount": map[string]interface{}{
				"amount":   "100.00",
				"currency": "USD",
			},
		}
		resp := ts.Post("/api/v1/budgets/"+budgetID+"/expected-expenses", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Equal(t, "No Category Expense", result["name"])
		_, hasCategoryID := result["category_id"]
		assert.False(t, hasCategoryID, "category_id should not be present when not set")
	})

	t.Run("DeleteExpectedExpense", func(t *testing.T) {
		resp := ts.Delete("/api/v1/expected-expenses/" + expectedID)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	// --- Actual Expenses ---
	var actualID string

	t.Run("CreateActualExpense", func(t *testing.T) {
		body := map[string]interface{}{
			"name":         "Supermarket",
			"description":  "Weekly groceries",
			"expense_date": time.Now().Format("2006-01-02"),
			"amount": map[string]interface{}{
				"amount":   "85.25",
				"currency": "ARS",
			},
		}
		resp := ts.Post("/api/v1/budgets/"+budgetID+"/actual-expenses", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Equal(t, "Supermarket", result["name"])
		amt := result["amount"].(map[string]interface{})
		assert.Equal(t, "85.25", amt["amount"])
		assert.Equal(t, "ARS", amt["currency"])
		actualID = result["id"].(string)
	})

	t.Run("GetActualExpense", func(t *testing.T) {
		resp := ts.Get("/api/v1/actual-expenses/" + actualID)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Equal(t, actualID, result["id"])
		assert.Equal(t, "Supermarket", result["name"])
	})

	t.Run("ListActualExpenses", func(t *testing.T) {
		resp := ts.Get("/api/v1/budgets/" + budgetID + "/actual-expenses")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.GreaterOrEqual(t, len(result), 1)
	})

	t.Run("UpdateActualExpense", func(t *testing.T) {
		body := map[string]interface{}{
			"name":         "Supermarket (Updated)",
			"description":  "Updated groceries",
			"expense_date": time.Now().Format("2006-01-02"),
			"amount": map[string]interface{}{
				"amount":   "92.50",
				"currency": "ARS",
			},
		}
		resp := ts.Put("/api/v1/actual-expenses/"+actualID, body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Equal(t, "Supermarket (Updated)", result["name"])
		amt := result["amount"].(map[string]interface{})
		assert.Equal(t, "92.5", amt["amount"])
	})

	t.Run("DeleteActualExpense", func(t *testing.T) {
		resp := ts.Delete("/api/v1/actual-expenses/" + actualID)
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})

	t.Run("CreateExpenseWithoutAuth", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Unauthorized",
			"amount": map[string]interface{}{
				"amount":   "10",
				"currency": "USD",
			},
		}
		resp := ts.DoRequest(http.MethodPost, "/api/v1/budgets/"+budgetID+"/expected-expenses", body, http.Header{})
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
