package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnboardingAPI(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	t.Run("GetDefaultOnboarding", func(t *testing.T) {
		resp := ts.Get("/api/v1/onboarding")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "in_progress", result["status"])
		assert.Equal(t, "welcome", result["current_step"])
		assert.Equal(t, []interface{}{}, result["steps"])
	})

	t.Run("CompleteWelcomeStep", func(t *testing.T) {
		resp := ts.Post("/api/v1/onboarding/steps/welcome/complete", map[string]interface{}{
			"data": map[string]interface{}{},
		})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "in_progress", result["status"])
		assert.Equal(t, "choose_group", result["current_step"])

		steps := result["steps"].([]interface{})
		require.Len(t, steps, 1)
		step := steps[0].(map[string]interface{})
		assert.Equal(t, "welcome", step["step"])
		assert.Equal(t, "completed", step["status"])
	})

	t.Run("CompleteChooseGroupStep", func(t *testing.T) {
		// First create a group
		groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
			"name":        "Onboarding Test Group",
			"description": "Test group for onboarding",
		})
		assert.Equal(t, http.StatusCreated, groupResp.Code)

		var groupResult map[string]interface{}
		err := json.Unmarshal(groupResp.Body.Bytes(), &groupResult)
		require.NoError(t, err)
		groupID := groupResult["id"].(string)

		// Complete the choose_group step with group_id
		resp := ts.Post("/api/v1/onboarding/steps/choose_group/complete", map[string]interface{}{
			"data": map[string]interface{}{
				"group_id": groupID,
			},
		})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "in_progress", result["status"])
		assert.Equal(t, "choose_cadence", result["current_step"])

		steps := result["steps"].([]interface{})
		require.Len(t, steps, 2)
		step := steps[1].(map[string]interface{})
		assert.Equal(t, "choose_group", step["step"])
		assert.Equal(t, "completed", step["status"])
		data := step["data"].(map[string]interface{})
		assert.Equal(t, groupID, data["group_id"])
	})

	t.Run("CompleteChooseGroupStep_InvalidGroupId", func(t *testing.T) {
		resp := ts.Post("/api/v1/onboarding/steps/choose_group/complete", map[string]interface{}{
			"data": map[string]interface{}{
				"group_id": "not-a-uuid",
			},
		})
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result["error"], "invalid_step_data")
	})

	t.Run("CompleteChooseGroupStep_MissingGroupId", func(t *testing.T) {
		resp := ts.Post("/api/v1/onboarding/steps/choose_group/complete", map[string]interface{}{
			"data": map[string]interface{}{},
		})
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("SkipStep", func(t *testing.T) {
		// Skip the choose_cadence step
		resp := ts.Post("/api/v1/onboarding/steps/choose_cadence/skip", map[string]interface{}{})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "in_progress", result["status"])
		assert.Equal(t, "add_expected_expenses", result["current_step"])

		// Find the skipped step
		steps := result["steps"].([]interface{})
		var skippedStep map[string]interface{}
		for _, s := range steps {
			step := s.(map[string]interface{})
			if step["step"] == "choose_cadence" {
				skippedStep = step
				break
			}
		}
		require.NotNil(t, skippedStep)
		assert.Equal(t, "skipped", skippedStep["status"])
	})

	t.Run("GoBackFromAddExpectedExpenses", func(t *testing.T) {
		// At this point current_step is add_expected_expenses (from SkipStep above)
		resp := ts.Post("/api/v1/onboarding/steps/add_expected_expenses/back", map[string]interface{}{})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "in_progress", result["status"])
		assert.Equal(t, "choose_cadence", result["current_step"])
	})

	t.Run("GoBackFromChooseCadence", func(t *testing.T) {
		// Now at choose_cadence, go back to choose_group
		resp := ts.Post("/api/v1/onboarding/steps/choose_cadence/back", map[string]interface{}{})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "choose_group", result["current_step"])
	})

	t.Run("GoBackFromWelcomeFails", func(t *testing.T) {
		// Reset to welcome first
		_ = ts.Post("/api/v1/onboarding/reset", map[string]interface{}{})

		resp := ts.Post("/api/v1/onboarding/steps/welcome/back", map[string]interface{}{})
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result["error"], "invalid_step_data")
	})

	t.Run("InvalidStep", func(t *testing.T) {
		resp := ts.Post("/api/v1/onboarding/steps/invalid_step/complete", map[string]interface{}{
			"data": map[string]interface{}{},
		})
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("ResetOnboarding", func(t *testing.T) {
		resp := ts.Post("/api/v1/onboarding/reset", map[string]interface{}{})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "in_progress", result["status"])
		assert.Equal(t, "welcome", result["current_step"])
		assert.Equal(t, []interface{}{}, result["steps"])
	})

	t.Run("CompleteAllSteps", func(t *testing.T) {
		// Reset first
		_ = ts.Post("/api/v1/onboarding/reset", map[string]interface{}{})

		// Complete welcome
		_ = ts.Post("/api/v1/onboarding/steps/welcome/complete", map[string]interface{}{
			"data": map[string]interface{}{},
		})

		// Create a group
		groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
			"name": "Complete Flow Group",
		})
		var groupResult map[string]interface{}
		_ = json.Unmarshal(groupResp.Body.Bytes(), &groupResult)
		groupID := groupResult["id"].(string)

		// Complete choose_group
		_ = ts.Post("/api/v1/onboarding/steps/choose_group/complete", map[string]interface{}{
			"data": map[string]interface{}{"group_id": groupID},
		})

		// Skip choose_cadence
		_ = ts.Post("/api/v1/onboarding/steps/choose_cadence/skip", map[string]interface{}{})

		// Skip add_expected_expenses
		_ = ts.Post("/api/v1/onboarding/steps/add_expected_expenses/skip", map[string]interface{}{})

		// Skip budget_summary
		_ = ts.Post("/api/v1/onboarding/steps/budget_summary/skip", map[string]interface{}{})

		// Skip register_actual_expense
		_ = ts.Post("/api/v1/onboarding/steps/register_actual_expense/skip", map[string]interface{}{})

		// Skip compare_expenses
		_ = ts.Post("/api/v1/onboarding/steps/compare_expenses/skip", map[string]interface{}{})

		// Skip dashboard_tour
		_ = ts.Post("/api/v1/onboarding/steps/dashboard_tour/skip", map[string]interface{}{})

		// Complete the final step
		resp := ts.Post("/api/v1/onboarding/steps/complete/complete", map[string]interface{}{
			"data": map[string]interface{}{},
		})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "completed", result["status"])
		assert.Equal(t, "complete", result["current_step"])
	})
}
