package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupAPI(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	var createdGroupID string

	t.Run("CreateGroup", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Test Family Budget",
			"description": "A test budgeting group",
		}

		resp := ts.Post("/api/v1/groups", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.NotEmpty(t, result["id"])
		assert.Equal(t, "Test Family Budget", result["name"])
		createdGroupID = result["id"].(string)
	})

	t.Run("GetGroup", func(t *testing.T) {
		resp := ts.Get("/api/v1/groups/" + createdGroupID)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, createdGroupID, result["id"])
		assert.Equal(t, "Test Family Budget", result["name"])
	})

	t.Run("ListGroups", func(t *testing.T) {
		resp := ts.Get("/api/v1/groups")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(result), 1)
	})

	t.Run("UpdateGroup", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "Updated Family Budget",
			"description": "Updated description",
		}

		resp := ts.Put("/api/v1/groups/"+createdGroupID, body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "Updated Family Budget", result["name"])
		assert.Equal(t, "Updated description", result["description"])
	})

	t.Run("DeleteGroup", func(t *testing.T) {
		resp := ts.Delete("/api/v1/groups/" + createdGroupID)
		assert.Equal(t, http.StatusNoContent, resp.Code)

		// Verify it's soft deleted (should return 404)
		resp = ts.Get("/api/v1/groups/" + createdGroupID)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("CreateGroupWithoutAuth", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Unauthorized Group",
		}

		resp := ts.DoRequest(http.MethodPost, "/api/v1/groups", body, http.Header{})
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
