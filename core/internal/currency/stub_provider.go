package currency

import (
	"context"
	"fmt"
	"time"

	"github.com/budgets/core/internal/domain"
	"github.com/shopspring/decimal"
)

// StubExchangeRateProvider is a stub provider for testing and development
type StubExchangeRateProvider struct {
	rates map[string]decimal.Decimal
}

func NewStubExchangeRateProvider() *StubExchangeRateProvider {
	// Default rates relative to USD
	rates := map[string]decimal.Decimal{
		"USD_USD": decimal.NewFromInt(1),
		"USD_EUR": decimal.NewFromFloat(0.92),
		"USD_GBP": decimal.NewFromFloat(0.79),
		"USD_ARS": decimal.NewFromFloat(1450.0),
		"USD_BRL": decimal.NewFromFloat(4.97),
		"USD_MXN": decimal.NewFromFloat(17.15),
		"USD_CLP": decimal.NewFromFloat(890.0),
		"USD_COP": decimal.NewFromFloat(3950.0),
		"USD_PEN": decimal.NewFromFloat(3.72),
		"USD_UYU": decimal.NewFromFloat(39.5),
		"EUR_USD": decimal.NewFromFloat(1.09),
		"GBP_USD": decimal.NewFromFloat(1.27),
		"ARS_USD": decimal.NewFromFloat(0.00118),
	}
	return &StubExchangeRateProvider{rates: rates}
}

func (p *StubExchangeRateProvider) ProviderName() string {
	return "stub"
}

func (p *StubExchangeRateProvider) GetRate(ctx context.Context, from, to domain.Currency) (*ExchangeRate, error) {
	key := fmt.Sprintf("%s_%s", from, to)
	
	if rate, ok := p.rates[key]; ok {
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         rate,
			Timestamp:    time.Now(),
			Source:       p.ProviderName(),
		}, nil
	}

	// Try to calculate through USD
	fromUSDKey := fmt.Sprintf("%s_USD", from)
	usdToKey := fmt.Sprintf("USD_%s", to)
	
	fromUSD, hasFromUSD := p.rates[fromUSDKey]
	usdTo, hasUSDTo := p.rates[usdToKey]
	
	if hasFromUSD && hasUSDTo {
		rate := fromUSD.Mul(usdTo)
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         rate,
			Timestamp:    time.Now(),
			Source:       p.ProviderName(),
		}, nil
	}

	return nil, fmt.Errorf("exchange rate not available for %s to %s", from, to)
}

func (p *StubExchangeRateProvider) GetRates(ctx context.Context, base domain.Currency, targets []domain.Currency) ([]ExchangeRate, error) {
	var rates []ExchangeRate
	for _, target := range targets {
		rate, err := p.GetRate(ctx, base, target)
		if err != nil {
			continue
		}
		rates = append(rates, *rate)
	}
	return rates, nil
}

func (p *StubExchangeRateProvider) GetHistoricalRate(ctx context.Context, from, to domain.Currency, date time.Time) (*ExchangeRate, error) {
	// Stub returns current rate for historical queries
	return p.GetRate(ctx, from, to)
}

func (p *StubExchangeRateProvider) SetRate(from, to domain.Currency, rate decimal.Decimal) {
	key := fmt.Sprintf("%s_%s", from, to)
	p.rates[key] = rate
}
