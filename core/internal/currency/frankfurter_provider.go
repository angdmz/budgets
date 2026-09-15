package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/budgets/core/internal/domain"
	"github.com/shopspring/decimal"
)

const frankfurterBaseURL = "https://api.frankfurter.dev/v2"

type FrankfurterProvider struct {
	client  *http.Client
	baseURL string
}

func NewFrankfurterProvider(timeoutSeconds int) *FrankfurterProvider {
	return &FrankfurterProvider{
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		baseURL: frankfurterBaseURL,
	}
}

type frankfurterRateEntry struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

func (p *FrankfurterProvider) ProviderName() string {
	return "frankfurter"
}

func (p *FrankfurterProvider) GetRate(ctx context.Context, from, to domain.Currency, quote domain.QuoteType) (*ExchangeRate, error) {
	if from == to {
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Quote:        quote,
			Rate:         decimal.NewFromInt(1),
			Timestamp:    time.Now(),
			Source:       p.ProviderName(),
		}, nil
	}

	url := fmt.Sprintf("%s/rates?base=%s&quotes=%s", p.baseURL, string(from), string(to))
	rate, err := p.fetchRate(ctx, url, from, to, quote)
	if err != nil {
		return nil, err
	}
	return rate, nil
}

func (p *FrankfurterProvider) GetRates(ctx context.Context, base domain.Currency, targets []domain.Currency, quote domain.QuoteType) ([]ExchangeRate, error) {
	symbols := make([]string, 0, len(targets))
	for _, t := range targets {
		if t != base {
			symbols = append(symbols, string(t))
		}
	}
	if len(symbols) == 0 {
		return []ExchangeRate{}, nil
	}

	url := fmt.Sprintf("%s/rates?base=%s&quotes=%s", p.baseURL, string(base), strings.Join(symbols, ","))
	resp, err := p.fetchRates(ctx, url, base, quote)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *FrankfurterProvider) GetHistoricalRate(ctx context.Context, from, to domain.Currency, quote domain.QuoteType, date time.Time) (*ExchangeRate, error) {
	if from == to {
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Quote:        quote,
			Rate:         decimal.NewFromInt(1),
			Timestamp:    date,
			Source:       p.ProviderName(),
		}, nil
	}

	url := fmt.Sprintf("%s/rates?base=%s&quotes=%s&date=%s", p.baseURL, string(from), string(to), date.Format("2006-01-02"))
	rate, err := p.fetchRate(ctx, url, from, to, quote)
	if err != nil {
		return nil, err
	}
	rate.Timestamp = date
	return rate, nil
}

func (p *FrankfurterProvider) fetchRate(ctx context.Context, url string, from, to domain.Currency, quote domain.QuoteType) (*ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("frankfurter: failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("frankfurter: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("frankfurter: API returned status %d: %s", resp.StatusCode, string(body))
	}

	var entries []frankfurterRateEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("frankfurter: failed to decode response: %w", err)
	}

	for _, e := range entries {
		if e.Quote == string(to) {
			return &ExchangeRate{
				FromCurrency: from,
				ToCurrency:   to,
				Quote:        quote,
				Rate:         decimal.NewFromFloat(e.Rate),
				Timestamp:    time.Now(),
				Source:       p.ProviderName(),
			}, nil
		}
	}

	return nil, fmt.Errorf("frankfurter: rate for %s not found in response", to)
}

func (p *FrankfurterProvider) fetchRates(ctx context.Context, url string, base domain.Currency, quote domain.QuoteType) ([]ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("frankfurter: failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("frankfurter: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("frankfurter: API returned status %d: %s", resp.StatusCode, string(body))
	}

	var entries []frankfurterRateEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("frankfurter: failed to decode response: %w", err)
	}

	rates := make([]ExchangeRate, 0, len(entries))
	for _, e := range entries {
		rates = append(rates, ExchangeRate{
			FromCurrency: base,
			ToCurrency:   domain.Currency(e.Quote),
			Quote:        quote,
			Rate:         decimal.NewFromFloat(e.Rate),
			Timestamp:    time.Now(),
			Source:       p.ProviderName(),
		})
	}

	return rates, nil
}
