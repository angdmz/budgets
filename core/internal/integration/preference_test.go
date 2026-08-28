package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreferenceAPI(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	t.Run("GetDefaultPreferences", func(t *testing.T) {
		resp := ts.Get("/api/v1/preferences")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "LIGHT", result["theme"])
		assert.Equal(t, "EN", result["language"])
		assert.Equal(t, "USD", result["display_currency"])
	})

	t.Run("UpdatePreferences_DarkThemeSpanish", func(t *testing.T) {
		body := map[string]interface{}{
			"theme":            "DARK",
			"language":         "ES",
			"display_currency": "ARS",
		}
		resp := ts.Put("/api/v1/preferences", body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "DARK", result["theme"])
		assert.Equal(t, "ES", result["language"])
		assert.Equal(t, "ARS", result["display_currency"])
	})

	t.Run("GetUpdatedPreferences", func(t *testing.T) {
		resp := ts.Get("/api/v1/preferences")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "DARK", result["theme"])
		assert.Equal(t, "ES", result["language"])
		assert.Equal(t, "ARS", result["display_currency"])
	})

	t.Run("UpdatePreferences_DimTheme", func(t *testing.T) {
		body := map[string]interface{}{
			"theme":            "DIM",
			"language":         "EN",
			"display_currency": "EUR",
		}
		resp := ts.Put("/api/v1/preferences", body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "DIM", result["theme"])
		assert.Equal(t, "EN", result["language"])
		assert.Equal(t, "EUR", result["display_currency"])
	})

	t.Run("UpdatePreferences_InvalidTheme", func(t *testing.T) {
		body := map[string]interface{}{
			"theme":            "NEON",
			"language":         "EN",
			"display_currency": "USD",
		}
		resp := ts.Put("/api/v1/preferences", body)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("UpdatePreferences_InvalidLanguage", func(t *testing.T) {
		body := map[string]interface{}{
			"theme":            "LIGHT",
			"language":         "FR",
			"display_currency": "USD",
		}
		resp := ts.Put("/api/v1/preferences", body)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("UpdatePreferences_InvalidCurrency", func(t *testing.T) {
		body := map[string]interface{}{
			"theme":            "LIGHT",
			"language":         "EN",
			"display_currency": "BTC",
		}
		resp := ts.Put("/api/v1/preferences", body)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("GetPreferencesWithoutAuth", func(t *testing.T) {
		resp := ts.DoRequest(http.MethodGet, "/api/v1/preferences", nil, http.Header{})
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("PatchPreferences_OnlyTheme", func(t *testing.T) {
		body := map[string]interface{}{
			"theme": "DARK",
		}
		resp := ts.DoRequest(http.MethodPatch, "/api/v1/preferences", body, ts.AuthHeader())
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "DARK", result["theme"])
		assert.Equal(t, "EN", result["language"])
		assert.Equal(t, "EUR", result["display_currency"])
	})

	t.Run("PatchPreferences_OnlyCurrency", func(t *testing.T) {
		body := map[string]interface{}{
			"display_currency": "BRL",
		}
		resp := ts.DoRequest(http.MethodPatch, "/api/v1/preferences", body, ts.AuthHeader())
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		assert.Equal(t, "DARK", result["theme"])
		assert.Equal(t, "EN", result["language"])
		assert.Equal(t, "BRL", result["display_currency"])
	})

	t.Run("PatchPreferences_EmptyBody", func(t *testing.T) {
		body := map[string]interface{}{}
		resp := ts.DoRequest(http.MethodPatch, "/api/v1/preferences", body, ts.AuthHeader())
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("PatchPreferences_InvalidTheme", func(t *testing.T) {
		body := map[string]interface{}{
			"theme": "NEON",
		}
		resp := ts.DoRequest(http.MethodPatch, "/api/v1/preferences", body, ts.AuthHeader())
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
