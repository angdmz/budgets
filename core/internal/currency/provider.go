package currency

import (
	"context"
	"time"

	"github.com/budgets/core/internal/domain"
	"github.com/shopspring/decimal"
)

// ExchangeRate represents an exchange rate between two currencies
type ExchangeRate struct {
	FromCurrency domain.Currency
	ToCurrency   domain.Currency
	Rate         decimal.Decimal
	Timestamp    time.Time
	Source       string
}

// ExchangeRateProvider is the interface for fetching exchange rates
type ExchangeRateProvider interface {
	// GetRate returns the current exchange rate between two currencies
	GetRate(ctx context.Context, from, to domain.Currency) (*ExchangeRate, error)

	// GetRates returns exchange rates for multiple currency pairs
	GetRates(ctx context.Context, base domain.Currency, targets []domain.Currency) ([]ExchangeRate, error)

	// GetHistoricalRate returns the exchange rate at a specific date
	GetHistoricalRate(ctx context.Context, from, to domain.Currency, date time.Time) (*ExchangeRate, error)

	// ProviderName returns the name of this provider
	ProviderName() string
}

// CurrencyMarketplace provides currency conversion services
type CurrencyMarketplace struct {
	provider ExchangeRateProvider
	cache    ExchangeRateCache
}

// ExchangeRateCache caches exchange rates
type ExchangeRateCache interface {
	Get(from, to domain.Currency) (*ExchangeRate, bool)
	Set(rate ExchangeRate, ttl time.Duration)
	Clear()
}

func NewCurrencyMarketplace(provider ExchangeRateProvider, cache ExchangeRateCache) *CurrencyMarketplace {
	return &CurrencyMarketplace{
		provider: provider,
		cache:    cache,
	}
}

// Convert converts an amount from one currency to another
func (m *CurrencyMarketplace) Convert(ctx context.Context, amount domain.Money, to domain.Currency) (domain.Money, error) {
	if amount.Currency == to {
		return amount, nil
	}

	// Check cache first
	if m.cache != nil {
		if rate, ok := m.cache.Get(amount.Currency, to); ok {
			convertedAmount := amount.Amount.Mul(rate.Rate)
			return domain.NewMoney(convertedAmount, to), nil
		}
	}

	// Fetch from provider
	rate, err := m.provider.GetRate(ctx, amount.Currency, to)
	if err != nil {
		return domain.Money{}, err
	}

	// Cache the rate
	if m.cache != nil {
		m.cache.Set(*rate, 5*time.Minute)
	}

	convertedAmount := amount.Amount.Mul(rate.Rate)
	return domain.NewMoney(convertedAmount, to), nil
}

// GetExchangeRate returns the exchange rate between two currencies
func (m *CurrencyMarketplace) GetExchangeRate(ctx context.Context, from, to domain.Currency) (*ExchangeRate, error) {
	if from == to {
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         decimal.NewFromInt(1),
			Timestamp:    time.Now(),
			Source:       "identity",
		}, nil
	}

	// Check cache first
	if m.cache != nil {
		if rate, ok := m.cache.Get(from, to); ok {
			return rate, nil
		}
	}

	rate, err := m.provider.GetRate(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Cache the rate
	if m.cache != nil {
		m.cache.Set(*rate, 5*time.Minute)
	}

	return rate, nil
}
