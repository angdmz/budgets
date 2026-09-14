package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/encryption"
)

// validateMoneyRequest parses and validates a MoneyRequest, returning the
// parsed decimal amount and validated currency. On failure it writes a 400
// response to the gin context and returns ok=false.
func validateMoneyRequest(c *gin.Context, req MoneyRequest) (decimal.Decimal, domain.Currency, bool) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Invalid amount format"})
		return decimal.Zero, "", false
	}

	currency := domain.Currency(req.Currency)
	if !currency.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_currency", Message: "Unsupported currency"})
		return decimal.Zero, "", false
	}

	return amount, currency, true
}

// parseExpenseDate parses a date string in YYYY-MM-DD format. On failure it
// writes a 400 response to the gin context and returns ok=false.
func parseExpenseDate(c *gin.Context, dateStr string) (time.Time, bool) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_expense_date", Message: "Date must be in YYYY-MM-DD format"})
		return time.Time{}, false
	}
	return date, true
}

// toExpectedExpenseResponse builds an ExpectedExpenseResponse from a persisted
// expected expense and its decrypted money.
func toExpectedExpenseResponse(expense *domain.PersistedExpectedExpense, decryptedMoney encryption.Money) ExpectedExpenseResponse {
	return ExpectedExpenseResponse{
		ID:          expense.ExternalID(),
		Name:        expense.Name(),
		Description: expense.Description(),
		Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
		CategoryID:  expense.CategoryExternalID(),
		CreatedAt:   expense.CreatedAt(),
		UpdatedAt:   expense.UpdatedAt(),
	}
}

// toActualExpenseResponse builds an ActualExpenseResponse from a persisted
// actual expense and its decrypted money.
func toActualExpenseResponse(expense *domain.PersistedActualExpense, decryptedMoney encryption.Money) ActualExpenseResponse {
	return ActualExpenseResponse{
		ID:          expense.ExternalID(),
		Name:        expense.Name(),
		Description: expense.Description(),
		ExpenseDate: expense.ExpenseDate().Format("2006-01-02"),
		Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
		CategoryID:  expense.CategoryExternalID(),
		CreatedAt:   expense.CreatedAt(),
		UpdatedAt:   expense.UpdatedAt(),
	}
}
