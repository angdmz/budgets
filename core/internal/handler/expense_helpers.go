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

// encryptorDecrypter adapts the encryption facility to the domain's
// AmountDecrypter contract for expense rendering.
type encryptorDecrypter struct {
	enc *encryption.Encryptor
}

func (d encryptorDecrypter) DecryptAmount(ciphertext string) (domain.Money, error) {
	m, err := d.enc.DecryptMoney(ciphertext)
	if err != nil {
		return domain.Money{}, err
	}
	return domain.NewMoney(m.Amount, domain.Currency(m.Currency)), nil
}
