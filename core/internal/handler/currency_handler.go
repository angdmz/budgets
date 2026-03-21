package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/budgets/core/internal/currency"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
)

type CurrencyHandler struct {
	marketplace *currency.CurrencyMarketplace
}

func NewCurrencyHandler(marketplace *currency.CurrencyMarketplace) *CurrencyHandler {
	return &CurrencyHandler{marketplace: marketplace}
}

// Convert godoc
// @Summary Convert an amount between currencies
// @Description Convert a monetary amount from one currency to another
// @Tags currency
// @Accept json
// @Produce json
// @Param request body ConvertCurrencyRequest true "Conversion request"
// @Success 200 {object} ConvertCurrencyResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /currency/convert [post]
func (h *CurrencyHandler) Convert(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var req ConvertCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount"})
		return
	}

	from := domain.Currency(req.FromCurrency)
	to := domain.Currency(req.ToCurrency)

	if !from.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_from_currency"})
		return
	}
	if !to.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_to_currency"})
		return
	}

	money := domain.NewMoney(amount, from)
	converted, err := h.marketplace.Convert(c.Request.Context(), money, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "conversion_failed", Message: err.Error()})
		return
	}

	rate, _ := h.marketplace.GetExchangeRate(c.Request.Context(), from, to)
	rateStr := ""
	provider := ""
	if rate != nil {
		rateStr = rate.Rate.String()
		provider = rate.Source
	}

	c.JSON(http.StatusOK, ConvertCurrencyResponse{
		OriginalAmount:  MoneyResponse{Amount: money.Amount.String(), Currency: string(money.Currency)},
		ConvertedAmount: MoneyResponse{Amount: converted.Amount.String(), Currency: string(converted.Currency)},
		ExchangeRate:    rateStr,
		Provider:        provider,
	})
}

// GetExchangeRates godoc
// @Summary Get exchange rates for a base currency
// @Description Get exchange rates from a base currency to all supported currencies
// @Tags currency
// @Produce json
// @Param base query string true "Base currency code (e.g. USD)"
// @Success 200 {array} ExchangeRateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /currency/rates [get]
func (h *CurrencyHandler) GetExchangeRates(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	base := domain.Currency(c.Query("base"))
	if !base.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_base_currency"})
		return
	}

	targets := []domain.Currency{
		domain.CurrencyUSD, domain.CurrencyEUR, domain.CurrencyGBP,
		domain.CurrencyARS, domain.CurrencyBRL, domain.CurrencyMXN,
		domain.CurrencyCLP, domain.CurrencyCOP, domain.CurrencyPEN,
		domain.CurrencyUYU,
	}

	var response []ExchangeRateResponse
	for _, target := range targets {
		if target == base {
			continue
		}
		rate, err := h.marketplace.GetExchangeRate(c.Request.Context(), base, target)
		if err != nil {
			continue
		}
		response = append(response, ExchangeRateResponse{
			FromCurrency: string(base),
			ToCurrency:   string(target),
			Rate:         rate.Rate.String(),
			Provider:     rate.Source,
		})
	}

	c.JSON(http.StatusOK, response)
}
