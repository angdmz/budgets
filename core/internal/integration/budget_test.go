package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBudgetAPI(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	// Setup: create a group
	var groupID string
	t.Run("SetupGroup", func(t *testing.T) {
		resp := ts.Post("/api/v1/groups", map[string]interface{}{
			"name": "Budget Test Group",
		})
		require.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		groupID = result["id"].(string)
	})

	var budgetID string
	now := time.Now()
	startDate := now.Format("2006-01-02")
	endDate := now.AddDate(0, 1, 0).Format("2006-01-02")

	t.Run("CreateBudget", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "March 2026",
			"description": "Monthly budget",
			"start_date":  startDate,
			"end_date":    endDate,
		}
		resp := ts.Post("/api/v1/groups/"+groupID+"/budgets", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "March 2026", result["name"])
		assert.Equal(t, startDate, result["start_date"])
		assert.Equal(t, endDate, result["end_date"])
		budgetID = result["id"].(string)
	})

	t.Run("GetBudget", func(t *testing.T) {
		resp := ts.Get("/api/v1/budgets/" + budgetID)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, budgetID, result["id"])
		assert.Equal(t, "March 2026", result["name"])
	})

	t.Run("ListBudgets", func(t *testing.T) {
		resp := ts.Get("/api/v1/groups/" + groupID + "/budgets")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
	})

	t.Run("UpdateBudget", func(t *testing.T) {
		newEnd := now.AddDate(0, 3, 0).Format("2006-01-02")
		body := map[string]interface{}{
			"name":        "Q1 2026",
			"description": "Quarterly budget",
			"start_date":  startDate,
			"end_date":    newEnd,
		}
		resp := ts.Put("/api/v1/budgets/"+budgetID, body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "Q1 2026", result["name"])
		assert.Equal(t, "Quarterly budget", result["description"])
		assert.Equal(t, newEnd, result["end_date"])
	})

	t.Run("DeleteBudget", func(t *testing.T) {
		resp := ts.Delete("/api/v1/budgets/" + budgetID)
		assert.Equal(t, http.StatusNoContent, resp.Code)

		resp = ts.Get("/api/v1/budgets/" + budgetID)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("CreateBudgetWithoutAuth", func(t *testing.T) {
		body := map[string]interface{}{
			"name":       "Unauthorized",
			"start_date": startDate,
			"end_date":   endDate,
		}
		resp := ts.DoRequest(http.MethodPost, "/api/v1/groups/"+groupID+"/budgets", body, http.Header{})
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
