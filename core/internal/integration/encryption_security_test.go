package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionSecurity(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	var groupID, budgetID, expenseID string

	t.Run("SetupTestData", func(t *testing.T) {
		// Create a group
		groupBody := map[string]interface{}{
			"name": "Encryption Test Group",
		}
		resp := ts.Post("/api/v1/groups", groupBody)
		require.Equal(t, http.StatusCreated, resp.Code)

		var groupResult map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &groupResult)
		groupID = groupResult["id"].(string)

		// Create a budget
		budgetBody := map[string]interface{}{
			"name":       "Test Budget",
			"start_date": "2024-01-01",
			"end_date":   "2024-01-31",
		}
		resp = ts.Post("/api/v1/groups/"+groupID+"/budgets", budgetBody)
		require.Equal(t, http.StatusCreated, resp.Code)

		var budgetResult map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &budgetResult)
		budgetID = budgetResult["id"].(string)

		// Create an actual expense with sensitive monetary data
		expenseBody := map[string]interface{}{
			"name":         "Secret Expense",
			"description":  "This contains sensitive financial data",
			"expense_date": "2024-01-15",
			"amount": map[string]interface{}{
				"amount":   "12345.67",
				"currency": "USD",
			},
		}
		resp = ts.Post("/api/v1/budgets/"+budgetID+"/actual-expenses", expenseBody)
		require.Equal(t, http.StatusCreated, resp.Code)

		var expenseResult map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &expenseResult)
		expenseID = expenseResult["id"].(string)
	})

	t.Run("VerifyEncryptedDataInDatabase", func(t *testing.T) {
		// Query the database directly to verify the amount is encrypted
		ctx := context.Background()

		var encryptedAmount string
		err := ts.DB.QueryRow(ctx, `
			SELECT encrypted_amount 
			FROM actual_expenses 
			WHERE external_id = $1
		`, expenseID).Scan(&encryptedAmount)

		require.NoError(t, err)
		require.NotEmpty(t, encryptedAmount)

		// Verify the stored value is NOT plaintext
		assert.NotContains(t, encryptedAmount, "12345.67",
			"Amount should be encrypted, not stored as plaintext")
		assert.NotContains(t, encryptedAmount, "USD",
			"Currency should be encrypted, not stored as plaintext")

		// Verify the encrypted value looks like Fernet token (base64 encoded)
		assert.True(t, len(encryptedAmount) > 50,
			"Encrypted value should be longer than plaintext due to Fernet overhead")

		t.Logf("Encrypted amount stored in DB: %s", encryptedAmount[:50]+"...")
	})

	t.Run("VerifyAPIReturnsDecryptedData", func(t *testing.T) {
		// Fetch the expense via API - should return decrypted data
		resp := ts.Get("/api/v1/actual-expenses/" + expenseID)
		require.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		// Verify the API returns the decrypted amount
		amount := result["amount"].(map[string]interface{})
		assert.Equal(t, "12345.67", amount["amount"],
			"API should return decrypted amount")
		assert.Equal(t, "USD", amount["currency"],
			"API should return decrypted currency")
	})

	t.Run("VerifyEncryptionWithDifferentAmounts", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple expenses with different amounts
		amounts := []string{"100.00", "200.00", "100.00"}
		var encryptedValues []string

		for i, amt := range amounts {
			expenseBody := map[string]interface{}{
				"name":         "Test Expense " + string(rune('A'+i)),
				"expense_date": "2024-01-15",
				"amount": map[string]interface{}{
					"amount":   amt,
					"currency": "USD",
				},
			}
			resp := ts.Post("/api/v1/budgets/"+budgetID+"/actual-expenses", expenseBody)
			require.Equal(t, http.StatusCreated, resp.Code)

			var expenseResult map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &expenseResult)
			expID := expenseResult["id"].(string)

			var encryptedAmount string
			err := ts.DB.QueryRow(ctx, `
				SELECT encrypted_amount 
				FROM actual_expenses 
				WHERE external_id = $1
			`, expID).Scan(&encryptedAmount)
			require.NoError(t, err)

			encryptedValues = append(encryptedValues, encryptedAmount)
		}

		// Verify that same plaintext produces different ciphertext (Fernet uses random IV)
		assert.NotEqual(t, encryptedValues[0], encryptedValues[2],
			"Same amount should produce different encrypted values due to random IV")

		// Verify all encrypted values are different from each other
		assert.NotEqual(t, encryptedValues[0], encryptedValues[1])
		assert.NotEqual(t, encryptedValues[1], encryptedValues[2])
	})

	t.Run("VerifyDirectDatabaseAccessCannotReadAmounts", func(t *testing.T) {
		ctx := context.Background()

		// Simulate an attacker querying the database directly
		rows, err := ts.DB.Query(ctx, `
			SELECT name, encrypted_amount 
			FROM actual_expenses 
			WHERE budget_id = (SELECT id FROM budgets WHERE external_id = $1)
		`, budgetID)
		require.NoError(t, err)
		defer rows.Close()

		var exposedData []struct {
			Name            string
			EncryptedAmount string
		}

		for rows.Next() {
			var name, encryptedAmount string
			err := rows.Scan(&name, &encryptedAmount)
			require.NoError(t, err)

			exposedData = append(exposedData, struct {
				Name            string
				EncryptedAmount string
			}{name, encryptedAmount})
		}

		// Verify attacker cannot extract meaningful financial data
		for _, data := range exposedData {
			// The encrypted amount should not contain any recognizable patterns
			assert.False(t, containsNumericPattern(data.EncryptedAmount),
				"Encrypted data should not contain recognizable numeric patterns")

			// The encrypted amount should not be decodable without the key
			assert.True(t, strings.HasPrefix(data.EncryptedAmount, "gAAAAA"),
				"Encrypted value should be a valid Fernet token")
		}

		t.Logf("Attacker would see %d expenses but cannot read amounts", len(exposedData))
	})
}

func containsNumericPattern(s string) bool {
	// Check if the string contains obvious numeric patterns like "123" or "100.00"
	patterns := []string{"123", "100", "200", "12345", ".00", ".67"}
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
