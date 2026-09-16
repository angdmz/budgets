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
var currencyNames = map[domain.Currency]string{
	domain.CurrencyUSD: "US Dollar",
	domain.CurrencyEUR: "Euro",
	domain.CurrencyGBP: "British Pound Sterling",
	domain.CurrencyARS: "Argentine Peso",
	domain.CurrencyBRL: "Brazilian Real",
	domain.CurrencyMXN: "Mexican Peso",
	domain.CurrencyCLP: "Chilean Peso",
	domain.CurrencyCOP: "Colombian Peso",
	domain.CurrencyPEN: "Peruvian Sol",
	domain.CurrencyUYU: "Uruguayan Peso",
}

// ListCurrencies godoc
// @Summary List all supported currencies
// @Description Returns the list of all supported currencies with their codes and names
// @Tags currency
// @Produce json
// @Success 200 {array} CurrencyResponse
// @Router /currencies [get]
func (h *CurrencyHandler) ListCurrencies(c *gin.Context) {
	currencies := domain.SupportedCurrencies()
	response := make([]CurrencyResponse, 0, len(currencies))
	for _, curr := range currencies {
		name := currencyNames[curr]
		response = append(response, CurrencyResponse{
			Code: string(curr),
			Name: name,
		})
	}
	c.JSON(http.StatusOK, response)
}

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
	quote := domain.QuoteType(req.QuoteType)
	if quote == "" {
		quote = domain.QuoteOfficial
	}

	if !from.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_from_currency"})
		return
	}
	if !to.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_to_currency"})
		return
	}
	if !quote.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_quote_type"})
		return
	}

	money := domain.NewMoney(amount, from)
	converted, err := h.marketplace.Convert(c.Request.Context(), money, to, quote)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "conversion_failed", Message: err.Error()})
		return
	}

	rate, _ := h.marketplace.GetExchangeRate(c.Request.Context(), from, to, quote)
	rateStr := ""
	provider := ""
	if rate != nil {
		rateStr = rate.Rate.String()
		provider = rate.Source
	}

	isConverted := money.Currency != converted.Currency
	c.JSON(http.StatusOK, ConvertCurrencyResponse{
		OriginalAmount:  MoneyResponse{Amount: money.Amount.String(), Currency: string(money.Currency)},
		ConvertedAmount: MoneyResponse{Amount: converted.Amount.String(), Currency: string(converted.Currency), Converted: isConverted},
		ExchangeRate:    rateStr,
		Provider:        provider,
		QuoteType:       string(quote),
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

	quote := domain.QuoteType(c.Query("quote_type"))
	if quote == "" {
		quote = domain.QuoteOfficial
	}
	if !quote.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_quote_type"})
		return
	}

	targets := domain.SupportedCurrencies()

	var response []ExchangeRateResponse
	for _, target := range targets {
		if target == base {
			continue
		}
		rate, err := h.marketplace.GetExchangeRate(c.Request.Context(), base, target, quote)
		if err != nil {
			continue
		}
		response = append(response, ExchangeRateResponse{
			FromCurrency: string(base),
			ToCurrency:   string(target),
			QuoteType:    string(quote),
			Rate:         rate.Rate.String(),
			Provider:     rate.Source,
		})
	}

	c.JSON(http.StatusOK, response)
}
