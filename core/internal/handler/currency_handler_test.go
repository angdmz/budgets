package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCurrencies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/currencies", nil)

	handler := &CurrencyHandler{}
	handler.ListCurrencies(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []CurrencyResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	assert.Len(t, result, 10)

	codes := make(map[string]bool)
	for _, curr := range result {
		codes[curr.Code] = true
		assert.NotEmpty(t, curr.Name)
	}

	for _, expected := range []string{"USD", "EUR", "GBP", "ARS", "BRL", "MXN", "CLP", "COP", "PEN", "UYU"} {
		assert.True(t, codes[expected], "expected currency %s in response", expected)
	}
}
