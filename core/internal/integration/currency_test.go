package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyAPI(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	t.Run("ConvertUSDtoEUR", func(t *testing.T) {
		body := map[string]interface{}{
			"amount":        "100.00",
			"from_currency": "USD",
			"to_currency":   "EUR",
		}
		resp := ts.Post("/api/v1/currency/convert", body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		orig := result["original_amount"].(map[string]interface{})
		assert.Equal(t, "100", orig["amount"])
		assert.Equal(t, "USD", orig["currency"])

		conv := result["converted_amount"].(map[string]interface{})
		assert.Equal(t, "EUR", conv["currency"])
		assert.NotEmpty(t, conv["amount"])
		assert.NotEmpty(t, result["exchange_rate"])
	})

	t.Run("ConvertSameCurrency", func(t *testing.T) {
		body := map[string]interface{}{
			"amount":        "50.00",
			"from_currency": "USD",
			"to_currency":   "USD",
		}
		resp := ts.Post("/api/v1/currency/convert", body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		conv := result["converted_amount"].(map[string]interface{})
		assert.Equal(t, "50", conv["amount"])
		assert.Equal(t, "USD", conv["currency"])
	})

	t.Run("ConvertInvalidCurrency", func(t *testing.T) {
		body := map[string]interface{}{
			"amount":        "100",
			"from_currency": "BTC",
			"to_currency":   "USD",
		}
		resp := ts.Post("/api/v1/currency/convert", body)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("GetExchangeRates", func(t *testing.T) {
		resp := ts.Get("/api/v1/currency/rates?base=USD")
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Greater(t, len(result), 0)

		for _, rate := range result {
			assert.Equal(t, "USD", rate["from_currency"])
			assert.NotEmpty(t, rate["to_currency"])
			assert.NotEmpty(t, rate["rate"])
		}
	})

	t.Run("GetExchangeRatesInvalidBase", func(t *testing.T) {
		resp := ts.Get("/api/v1/currency/rates?base=INVALID")
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("ConvertWithoutAuth", func(t *testing.T) {
		body := map[string]interface{}{
			"amount":        "100",
			"from_currency": "USD",
			"to_currency":   "EUR",
		}
		resp := ts.DoRequest(http.MethodPost, "/api/v1/currency/convert", body, http.Header{})
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("ListCurrencies", func(t *testing.T) {
		resp := ts.DoRequest(http.MethodGet, "/api/v1/currencies", nil, http.Header{})
		assert.Equal(t, http.StatusOK, resp.Code)

		var result []map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &result))
		assert.Len(t, result, 10)

		codes := make(map[string]bool)
		for _, c := range result {
			code := c["code"].(string)
			codes[code] = true
			assert.NotEmpty(t, c["name"])
		}

		for _, expected := range []string{"USD", "EUR", "GBP", "ARS", "BRL", "MXN", "CLP", "COP", "PEN", "UYU"} {
			assert.True(t, codes[expected], "expected currency %s in response", expected)
		}
	})
}
