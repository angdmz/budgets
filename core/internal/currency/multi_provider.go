package currency

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/budgets/core/internal/domain"
)

type MultiProvider struct {
	providers []ExchangeRateProvider
}

func NewMultiProvider(providers ...ExchangeRateProvider) *MultiProvider {
	return &MultiProvider{providers: providers}
}

func (m *MultiProvider) ProviderName() string {
	return "multi"
}

func (m *MultiProvider) GetRate(ctx context.Context, from, to domain.Currency, quote domain.QuoteType) (*ExchangeRate, error) {
	for _, p := range m.providers {
		rate, err := p.GetRate(ctx, from, to, quote)
		if err == nil {
			return rate, nil
		}
		log.Printf("multi-provider: %s failed for %s→%s (%s): %v", p.ProviderName(), from, to, quote, err)
	}
	return nil, fmt.Errorf("no provider available for %s→%s (%s)", from, to, quote)
}

func (m *MultiProvider) GetRates(ctx context.Context, base domain.Currency, targets []domain.Currency, quote domain.QuoteType) ([]ExchangeRate, error) {
	for _, p := range m.providers {
		rates, err := p.GetRates(ctx, base, targets, quote)
		if err == nil {
			return rates, nil
		}
		log.Printf("multi-provider: %s failed for GetRates base=%s: %v", p.ProviderName(), base, err)
	}
	return nil, fmt.Errorf("no provider available for rates with base %s", base)
}

func (m *MultiProvider) GetHistoricalRate(ctx context.Context, from, to domain.Currency, quote domain.QuoteType, date time.Time) (*ExchangeRate, error) {
	for _, p := range m.providers {
		rate, err := p.GetHistoricalRate(ctx, from, to, quote, date)
		if err == nil {
			return rate, nil
		}
		log.Printf("multi-provider: %s failed for historical %s→%s (%s) at %s: %v", p.ProviderName(), from, to, quote, date.Format("2006-01-02"), err)
	}
	return nil, fmt.Errorf("no provider available for historical %s→%s (%s) at %s", from, to, quote, date.Format("2006-01-02"))
}
