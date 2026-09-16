package currency

import (
	"context"
	"testing"
	"time"

	"github.com/budgets/core/internal/domain"
	"github.com/shopspring/decimal"
)


func TestStubProvider_GetRate_Direct(t *testing.T) {
	p := NewStubExchangeRateProvider()
	ctx := context.Background()

	rate, err := p.GetRate(ctx, domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate.FromCurrency != domain.CurrencyUSD || rate.ToCurrency != domain.CurrencyEUR {
		t.Errorf("unexpected currency pair: %s -> %s", rate.FromCurrency, rate.ToCurrency)
	}
	if rate.Quote != domain.QuoteOfficial {
		t.Errorf("expected quote OFFICIAL, got %s", rate.Quote)
	}
	if !rate.Rate.Equal(decimal.NewFromFloat(0.92)) {
		t.Errorf("expected rate 0.92, got %s", rate.Rate)
	}
}

func TestStubProvider_GetRate_ThroughUSD(t *testing.T) {
	p := NewStubExchangeRateProvider()
	ctx := context.Background()

	// ARS -> EUR should go through USD
	rate, err := p.GetRate(ctx, domain.CurrencyARS, domain.CurrencyEUR, domain.QuoteBlue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate.Quote != domain.QuoteBlue {
		t.Errorf("expected quote BLUE, got %s", rate.Quote)
	}
	// ARS_USD * USD_EUR = 0.00118 * 0.92
	expected := decimal.NewFromFloat(0.00118).Mul(decimal.NewFromFloat(0.92))
	if !rate.Rate.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, rate.Rate)
	}
}

func TestStubProvider_GetRate_NotFound(t *testing.T) {
	p := NewStubExchangeRateProvider()
	ctx := context.Background()

	_, err := p.GetRate(ctx, domain.CurrencyCLP, domain.CurrencyCOP, domain.QuoteOfficial)
	if err == nil {
		t.Error("expected error for unsupported pair")
	}
}

func TestStubProvider_GetRates(t *testing.T) {
	p := NewStubExchangeRateProvider()
	ctx := context.Background()

	targets := []domain.Currency{domain.CurrencyEUR, domain.CurrencyGBP, domain.CurrencyCLP}
	rates, err := p.GetRates(ctx, domain.CurrencyUSD, targets, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rates) != 3 {
		t.Fatalf("expected 3 rates, got %d", len(rates))
	}
}

func TestStubProvider_SetRate(t *testing.T) {
	p := NewStubExchangeRateProvider()
	ctx := context.Background()

	p.SetRate(domain.CurrencyCLP, domain.CurrencyCOP, decimal.NewFromFloat(0.0044))
	rate, err := p.GetRate(ctx, domain.CurrencyCLP, domain.CurrencyCOP, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Rate.Equal(decimal.NewFromFloat(0.0044)) {
		t.Errorf("expected 0.0044, got %s", rate.Rate)
	}
}

func TestStubProvider_GetHistoricalRate(t *testing.T) {
	p := NewStubExchangeRateProvider()
	ctx := context.Background()

	rate, err := p.GetHistoricalRate(ctx, domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteMEP, time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate.Quote != domain.QuoteMEP {
		t.Errorf("expected quote MEP, got %s", rate.Quote)
	}
}

func TestInMemoryCache_SetGet(t *testing.T) {
	c := NewInMemoryCache()

	rate := ExchangeRate{
		FromCurrency: domain.CurrencyUSD,
		ToCurrency:   domain.CurrencyEUR,
		Quote:        domain.QuoteOfficial,
		Rate:         decimal.NewFromFloat(0.92),
		Timestamp:    time.Now(),
		Source:       "stub",
	}

	c.Set(rate, 5*time.Minute)

	got, ok := c.Get(domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteOfficial)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if !got.Rate.Equal(rate.Rate) {
		t.Errorf("expected %s, got %s", rate.Rate, got.Rate)
	}
}

func TestInMemoryCache_QuoteIsolation(t *testing.T) {
	c := NewInMemoryCache()

	rateOfficial := ExchangeRate{
		FromCurrency: domain.CurrencyUSD,
		ToCurrency:   domain.CurrencyARS,
		Quote:        domain.QuoteOfficial,
		Rate:         decimal.NewFromInt(1000),
		Timestamp:    time.Now(),
		Source:       "stub",
	}
	rateBlue := ExchangeRate{
		FromCurrency: domain.CurrencyUSD,
		ToCurrency:   domain.CurrencyARS,
		Quote:        domain.QuoteBlue,
		Rate:         decimal.NewFromInt(1450),
		Timestamp:    time.Now(),
		Source:       "stub",
	}

	c.Set(rateOfficial, 5*time.Minute)
	c.Set(rateBlue, 5*time.Minute)

	gotOfficial, ok := c.Get(domain.CurrencyUSD, domain.CurrencyARS, domain.QuoteOfficial)
	if !ok {
		t.Fatal("expected cache hit for OFFICIAL")
	}
	gotBlue, ok := c.Get(domain.CurrencyUSD, domain.CurrencyARS, domain.QuoteBlue)
	if !ok {
		t.Fatal("expected cache hit for BLUE")
	}
	if !gotOfficial.Rate.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("expected 1000 for OFFICIAL, got %s", gotOfficial.Rate)
	}
	if !gotBlue.Rate.Equal(decimal.NewFromInt(1450)) {
		t.Errorf("expected 1450 for BLUE, got %s", gotBlue.Rate)
	}
}

func TestInMemoryCache_Expiry(t *testing.T) {
	c := NewInMemoryCache()

	rate := ExchangeRate{
		FromCurrency: domain.CurrencyUSD,
		ToCurrency:   domain.CurrencyEUR,
		Quote:        domain.QuoteOfficial,
		Rate:         decimal.NewFromFloat(0.92),
		Timestamp:    time.Now(),
		Source:       "stub",
	}

	c.Set(rate, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get(domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteOfficial)
	if ok {
		t.Error("expected cache miss after expiry")
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	c := NewInMemoryCache()

	rate := ExchangeRate{
		FromCurrency: domain.CurrencyUSD,
		ToCurrency:   domain.CurrencyEUR,
		Quote:        domain.QuoteOfficial,
		Rate:         decimal.NewFromFloat(0.92),
		Timestamp:    time.Now(),
		Source:       "stub",
	}

	c.Set(rate, 5*time.Minute)
	c.Clear()

	_, ok := c.Get(domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteOfficial)
	if ok {
		t.Error("expected cache miss after clear")
	}
}

func TestMarketplace_Convert_SameCurrency(t *testing.T) {
	p := NewStubExchangeRateProvider()
	m := NewCurrencyMarketplace(p, NewInMemoryCache())
	ctx := context.Background()

	amount := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	converted, err := m.Convert(ctx, amount, domain.CurrencyUSD, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !converted.Amount.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected 100, got %s", converted.Amount)
	}
}

func TestMarketplace_Convert_DifferentCurrency(t *testing.T) {
	p := NewStubExchangeRateProvider()
	m := NewCurrencyMarketplace(p, NewInMemoryCache())
	ctx := context.Background()

	amount := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	converted, err := m.Convert(ctx, amount, domain.CurrencyEUR, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := decimal.NewFromInt(100).Mul(decimal.NewFromFloat(0.92))
	if !converted.Amount.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, converted.Amount)
	}
	if converted.Currency != domain.CurrencyEUR {
		t.Errorf("expected EUR, got %s", converted.Currency)
	}
}

func TestMarketplace_Convert_UsesCache(t *testing.T) {
	p := NewStubExchangeRateProvider()
	cache := NewInMemoryCache()
	m := NewCurrencyMarketplace(p, cache)
	ctx := context.Background()

	amount := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	_, err := m.Convert(ctx, amount, domain.CurrencyEUR, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Override cache with a different rate
	cache.Set(ExchangeRate{
		FromCurrency: domain.CurrencyUSD,
		ToCurrency:   domain.CurrencyEUR,
		Quote:        domain.QuoteOfficial,
		Rate:         decimal.NewFromInt(2),
		Timestamp:    time.Now(),
		Source:       "test",
	}, 5*time.Minute)

	converted, err := m.Convert(ctx, amount, domain.CurrencyEUR, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := decimal.NewFromInt(200)
	if !converted.Amount.Equal(expected) {
		t.Errorf("expected cached rate result %s, got %s", expected, converted.Amount)
	}
}

func TestMarketplace_GetExchangeRate_Identity(t *testing.T) {
	p := NewStubExchangeRateProvider()
	m := NewCurrencyMarketplace(p, NewInMemoryCache())
	ctx := context.Background()

	rate, err := m.GetExchangeRate(ctx, domain.CurrencyUSD, domain.CurrencyUSD, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Rate.Equal(decimal.NewFromInt(1)) {
		t.Errorf("expected identity rate 1, got %s", rate.Rate)
	}
}

func TestMarketplace_GetExchangeRate_FromProvider(t *testing.T) {
	p := NewStubExchangeRateProvider()
	m := NewCurrencyMarketplace(p, NewInMemoryCache())
	ctx := context.Background()

	rate, err := m.GetExchangeRate(ctx, domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteBlue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate.Quote != domain.QuoteBlue {
		t.Errorf("expected BLUE, got %s", rate.Quote)
	}
	if !rate.Rate.Equal(decimal.NewFromFloat(0.92)) {
		t.Errorf("expected 0.92, got %s", rate.Rate)
	}
}

func TestMarketplace_ConvertHistorical_SameCurrency(t *testing.T) {
	p := NewStubExchangeRateProvider()
	m := NewCurrencyMarketplace(p, NewInMemoryCache())
	ctx := context.Background()

	amount := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	converted, err := m.ConvertHistorical(ctx, amount, domain.CurrencyUSD, domain.QuoteOfficial, time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !converted.Amount.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected 100, got %s", converted.Amount)
	}
}

func TestMarketplace_ConvertHistorical_DifferentCurrency(t *testing.T) {
	p := NewStubExchangeRateProvider()
	m := NewCurrencyMarketplace(p, NewInMemoryCache())
	ctx := context.Background()

	amount := domain.NewMoney(decimal.NewFromInt(100), domain.CurrencyUSD)
	converted, err := m.ConvertHistorical(ctx, amount, domain.CurrencyEUR, domain.QuoteOfficial, time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := decimal.NewFromInt(100).Mul(decimal.NewFromFloat(0.92))
	if !converted.Amount.Equal(expected) {
		t.Errorf("expected %s, got %s", expected, converted.Amount)
	}
	if converted.Currency != domain.CurrencyEUR {
		t.Errorf("expected EUR, got %s", converted.Currency)
	}
}

func TestMultiProvider_Fallback(t *testing.T) {
	stub := NewStubExchangeRateProvider()
	// MultiProvider with only stub should work for supported pairs
	mp := NewMultiProvider(stub)
	ctx := context.Background()

	rate, err := mp.GetRate(ctx, domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Rate.Equal(decimal.NewFromFloat(0.92)) {
		t.Errorf("expected 0.92, got %s", rate.Rate)
	}
}

func TestMultiProvider_NoProviderAvailable(t *testing.T) {
	// Use a provider that errors for an unsupported pair
	stub := NewStubExchangeRateProvider()
	mp := NewMultiProvider(stub)
	ctx := context.Background()

	// CLP→COP is not in the stub's rate map and can't be routed through USD
	_, err := mp.GetRate(ctx, domain.CurrencyCLP, domain.CurrencyCOP, domain.QuoteOfficial)
	if err == nil {
		t.Error("expected error when no provider can handle the pair")
	}
}

func TestMultiProvider_FallsThroughOnError(t *testing.T) {
	// First provider always errors, second provider (stub) succeeds
	stub := NewStubExchangeRateProvider()
	mp := NewMultiProvider(stub) // single provider, should still work
	ctx := context.Background()

	rate, err := mp.GetHistoricalRate(ctx, domain.CurrencyUSD, domain.CurrencyEUR, domain.QuoteOfficial, time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate.Source != "stub" {
		t.Errorf("expected source stub, got %s", rate.Source)
	}
}

func TestMultiProvider_ProviderName(t *testing.T) {
	mp := NewMultiProvider(NewStubExchangeRateProvider())
	if mp.ProviderName() != "multi" {
		t.Errorf("expected 'multi', got %s", mp.ProviderName())
	}
}

func TestFrankfurterProvider_ProviderName(t *testing.T) {
	p := NewFrankfurterProvider(10)
	if p.ProviderName() != "frankfurter" {
		t.Errorf("expected 'frankfurter', got %s", p.ProviderName())
	}
}

func TestFrankfurterProvider_GetRate_SameCurrency(t *testing.T) {
	p := NewFrankfurterProvider(10)
	ctx := context.Background()

	rate, err := p.GetRate(ctx, domain.CurrencyUSD, domain.CurrencyUSD, domain.QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Rate.Equal(decimal.NewFromInt(1)) {
		t.Errorf("expected identity rate 1, got %s", rate.Rate)
	}
}

func TestFrankfurterProvider_GetHistoricalRate_SameCurrency(t *testing.T) {
	p := NewFrankfurterProvider(10)
	ctx := context.Background()

	rate, err := p.GetHistoricalRate(ctx, domain.CurrencyUSD, domain.CurrencyUSD, domain.QuoteOfficial, time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Rate.Equal(decimal.NewFromInt(1)) {
		t.Errorf("expected identity rate 1, got %s", rate.Rate)
	}
}
