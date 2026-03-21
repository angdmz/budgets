package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryAPI(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	// First create a group to hold categories
	var groupID string
	t.Run("SetupGroup", func(t *testing.T) {
		resp := ts.Post("/api/v1/groups", map[string]interface{}{
			"name":        "Category Test Group",
			"description": "Group for category testing",
		})
		require.Equal(t, http.StatusCreated, resp.Code)
		var result map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		groupID = result["id"].(string)
	})

	var categoryID string

	t.Run("CreateCategory", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Groceries",
			"description": "Food and household items",
			"color":       "#4CAF50",
			"icon":        "cart",
		}
		resp := ts.Post("/api/v1/groups/"+groupID+"/categories", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "Groceries", result["name"])
		assert.Equal(t, "Food and household items", result["description"])
		assert.Equal(t, "#4CAF50", result["color"])
		assert.Equal(t, "cart", result["icon"])
		categoryID = result["id"].(string)
	})

	t.Run("ListCategories", func(t *testing.T) {
		resp := ts.Get("/api/v1/groups/" + groupID + "/categories")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)

		found := false
		for _, cat := range result {
			if cat["id"] == categoryID {
				found = true
				assert.Equal(t, "Groceries", cat["name"])
			}
		}
		assert.True(t, found, "Created category should be in list")
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Food & Groceries",
			"description": "Updated description",
			"color":       "#FF5722",
			"icon":        "food",
		}
		resp := ts.Put("/api/v1/categories/"+categoryID, body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "Food & Groceries", result["name"])
		assert.Equal(t, "Updated description", result["description"])
		assert.Equal(t, "#FF5722", result["color"])
		assert.Equal(t, "food", result["icon"])
	})

	t.Run("DeleteCategory", func(t *testing.T) {
		resp := ts.Delete("/api/v1/categories/" + categoryID)
		assert.Equal(t, http.StatusNoContent, resp.Code)

		// Verify it's soft deleted - list should not include it
		resp = ts.Get("/api/v1/groups/" + groupID + "/categories")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)
		for _, cat := range result {
			assert.NotEqual(t, categoryID, cat["id"], "Deleted category should not appear in list")
		}
	})

	t.Run("CreateCategoryWithoutAuth", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Unauthorized Category",
		}
		resp := ts.DoRequest(http.MethodPost, "/api/v1/groups/"+groupID+"/categories", body, http.Header{})
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
