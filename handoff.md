# Handoff: Currency Conversion — Multi-Provider + Backend Aggregation

## Context

The app needs currency conversion for budget summaries (totals, differences) displayed in the UI. Expenses can be in different currencies (USD, EUR, ARS, BRL, etc.), and aggregations must convert all amounts to the user's preferred `display_currency` using the user's `preferred_quote_type` before summing. **All conversion happens in the backend.**

## Previous Work (Completed)

Modal migration, UI refactor, responsive layout, accessibility audit — all done. See git history for details. Key artifacts still relevant:
- `app/src/components/BudgetSummary.tsx` — 3-card summary (expected, actual, difference) with `formatCurrency`. No changes needed.
- `app/src/hooks/useBudgetSelection.ts` — encapsulates group/budget/expense queries + derived totals. **Needs updating** (see Phase 3).
- `app/src/components/ExpenseList.tsx` — generic expense table. No changes needed.

## Current State

### What exists
- **Backend**: `CurrencyMarketplace` with `ExchangeRateProvider` interface, `InMemoryCache`, and `StubExchangeRateProvider` (hardcoded rates). A `POST /currency/convert` and `GET /currency/rates` endpoint exist but are not called by the frontend.
- **Backend**: `GET /budgets/:budget_id/summary` endpoint exists (`expense_handler.go:713-786`) but **naively sums amounts regardless of currency** — no conversion.
- **Frontend**: User preferences (`display_currency`, `preferred_quote_type`) are saved via `PATCH /preferences` and rendered in `Layout.tsx`, but **never used for conversion**.
- **Frontend**: `useBudgetSelection.ts` and `BudgetDetail.tsx` both do client-side `parseFloat` sums across mixed currencies — incorrect.

### What doesn't exist
- Real exchange rate providers (only stub).
- Any wiring of `CurrencyMarketplace` into `ExpenseHandler`.
- Any frontend call to the summary endpoint or convert endpoint.

## Decisions Made

### 1. All conversion in the backend
The frontend should call `GET /budgets/:budget_id/summary` and display the returned totals. No client-side conversion.

### 2. Historical rates for actual expenses
Actual expenses use the exchange rate at their `expense_date` (historical). Expected expenses use the budget's `end_date` (or current date if budget hasn't ended).

### 3. Multi-provider architecture (chain-of-responsibility)
`MultiProvider` wraps multiple `ExchangeRateProvider` implementations in priority order. `GetRate` tries each child in order; if one returns an error (unsupported pair/quote), it falls through to the next. **No `Supports` method on the interface** — routing is internal to `MultiProvider`.

### 4. `Supports` removed from `ExchangeRateProvider` interface
`Supports(quote domain.QuoteType) bool` was deemed an internal routing concern and should NOT belong to the public interface. The interface should only define capability (`GetRate`, `GetRates`, `GetHistoricalRate`, `ProviderName`).

### 5. Frankfurter for now, DolarAPI later
- **Frankfurter** (`api.frankfurter.dev/v2/rates`) — free, no API key, supports historical rates via `?date=YYYY-MM-DD`. Handles all currencies with `OFFICIAL` quote type. This is the only real provider to implement now.
- **DolarAPI** (`dolarapi.com/v1/dolares`) — for ARS non-official quotes (BLUE, MEP, CCL, CRYPTO) in the future. Returns `{compra, venta, casa, moneda, fechaActualizacion}`. Historical ARS rates via `argentinadatos.com/api/v1/cotizaciones/dolares/{casa}/{YYYY/MM/DD}`. **Not implemented now** — just keep the architecture extensible.

## Implementation Plan

### Phase 1: Provider Architecture (backend)

#### 1a. Update `ExchangeRateProvider` interface
**File**: `core/internal/currency/provider.go`

Remove `Supports(quote domain.QuoteType) bool` from the interface. Final interface:
```go
type ExchangeRateProvider interface {
    GetRate(ctx context.Context, from, to domain.Currency, quote domain.QuoteType) (*ExchangeRate, error)
    GetRates(ctx context.Context, base domain.Currency, targets []domain.Currency, quote domain.QuoteType) ([]ExchangeRate, error)
    GetHistoricalRate(ctx context.Context, from, to domain.Currency, quote domain.QuoteType, date time.Time) (*ExchangeRate, error)
    ProviderName() string
}
```

#### 1b. Update `StubExchangeRateProvider`
**File**: `core/internal/currency/stub_provider.go`

Remove the `Supports` method. Everything else stays.

#### 1c. Create `FrankfurterProvider`
**File**: `core/internal/currency/frankfurter_provider.go` (new)

- Implements `ExchangeRateProvider`.
- `GetRate`: calls `https://api.frankfurter.dev/v2/rates?base={from}&quotes={to}`. Parses response, returns `ExchangeRate` with `Quote: quote` (only meaningful for `OFFICIAL`; for non-official quotes, returns the official rate since Frankfurter doesn't distinguish).
- `GetRates`: calls `https://api.frankfurter.dev/v2/rates?base={base}&quotes={target1,target2,...}`.
- `GetHistoricalRate`: calls `https://api.frankfurter.dev/v2/rates?base={from}&quotes={to}&date={YYYY-MM-DD}`.
- `ProviderName`: returns `"frankfurter"`.
- Uses `net/http` client with configurable timeout from `ExchangeConfig.TimeoutSeconds()`.
- Returns error for unsupported currency pairs (Frankfurter supports ~30 currencies from ECB; all 10 supported currencies in this app should be covered).

#### 1d. Create `MultiProvider`
**File**: `core/internal/currency/multi_provider.go` (new)

```go
type MultiProvider struct {
    providers []ExchangeRateProvider
}

func NewMultiProvider(providers ...ExchangeRateProvider) *MultiProvider {
    return &MultiProvider{providers: providers}
}

func (m *MultiProvider) GetRate(ctx context.Context, from, to domain.Currency, quote domain.QuoteType) (*ExchangeRate, error) {
    for _, p := range m.providers {
        rate, err := p.GetRate(ctx, from, to, quote)
        if err == nil {
            return rate, nil
        }
        // log error, try next provider
    }
    return nil, fmt.Errorf("no provider available for %s→%s (%s)", from, to, quote)
}
// Same pattern for GetRates and GetHistoricalRate.
// ProviderName returns "multi".
```

**Provider order matters**: future DolarAPI goes first (claims ARS non-official), Frankfurter second (handles everything else). For now, only Frankfurter (or stub) is in the list.

#### 1e. Wire providers in `server.go`
**File**: `core/internal/server/server.go` (lines 34-57)

Based on `cfg.Exchange.Provider`:
- `"stub"` → `NewStubExchangeRateProvider()` (dev/test default)
- `"frankfurter"` → `NewFrankfurterProvider(cfg.Exchange.TimeoutSeconds())`
- Future: `"multi"` → `NewMultiProvider(dolarAPIProvider, frankfurterProvider)`

Pass `marketplace` to `NewExpenseHandler(pool, enc, marketplace)` (currently only `CurrencyHandler` gets it).

### Phase 2: Backend Conversion in Summary

#### 2a. Inject `CurrencyMarketplace` into `ExpenseHandler`
**File**: `core/internal/handler/expense_handler.go`

Add `marketplace *currency.CurrencyMarketplace` field to `ExpenseHandler` struct. Update `NewExpenseHandler` signature to accept it.

#### 2b. Modify `GetBudgetSummary`
**File**: `core/internal/handler/expense_handler.go` (lines 713-786)

Inside the `WithPersister` block:
1. Load user preferences: `domain.PersistedUserPreferenceFromPersistence(ctx, user.ID, p)` → get `displayCurrency` and `preferredQuoteType`.
2. For each **actual expense**: decrypt money → `marketplace.GetExchangeRate(ctx, money.Currency, displayCurrency, preferredQuoteType)` using `expense.ExpenseDate()` for historical rate → multiply → add to total.
3. For each **expected expense**: decrypt money → convert using budget's `end_date` (or now) → add to total.
4. Return `BudgetSummaryResponse` with `Currency: displayCurrency`, `Converted: true` on all `MoneyResponse` fields.

**Note**: `CurrencyMarketplace.Convert` currently doesn't support historical dates. Either:
- Add a `ConvertHistorical(ctx, amount, to, quote, date)` method to `CurrencyMarketplace`, or
- Extend `Convert` with an optional date parameter.
The marketplace should check cache, then call `provider.GetHistoricalRate` instead of `provider.GetRate` when a date is provided.

#### 2c. Update `CurrencyMarketplace` for historical conversion
**File**: `core/internal/currency/provider.go`

Add method:
```go
func (m *CurrencyMarketplace) ConvertHistorical(ctx context.Context, amount domain.Money, to domain.Currency, quote domain.QuoteType, date time.Time) (domain.Money, error)
```
Same as `Convert` but calls `m.provider.GetHistoricalRate` instead of `m.provider.GetRate`. Cache key should include the date (or skip cache for historical rates since they don't change).

### Phase 3: Frontend Changes

#### 3a. Add `BudgetSummary` type
**File**: `app/src/lib/types.ts`

```ts
export interface BudgetSummary {
  budget_id: string;
  expected_total: Money;
  actual_total: Money;
  difference: Money;
}
```

#### 3b. Replace naive sums in `useBudgetSelection.ts`
**File**: `app/src/hooks/useBudgetSelection.ts` (lines 54-57)

Replace the `reduce`/`parseFloat` logic with a `useQuery` calling `GET /budgets/:budget_id/summary`. Return `expectedTotal`, `actualTotal`, `difference`, `currency` from the response.

#### 3c. Replace naive sums in `BudgetDetail.tsx`
**File**: `app/src/pages/BudgetDetail.tsx` (lines 227-229)

Same: call `GET /budgets/:budget_id/summary` instead of client-side summing.

#### 3d. `BudgetSummary.tsx` — no changes needed
Already takes `expectedTotal`, `actualTotal`, `difference`, `currency` as props. Will just receive the backend-converted values now.

### Phase 4: Tests

#### 4a. Provider tests
**File**: `core/internal/currency/currency_test.go`

- Test `FrankfurterProvider` with an HTTP mock (or integration test against the real API with a timeout).
- Test `MultiProvider` routing: first provider errors → falls through to second.
- Test `MultiProvider` with no matching provider → error.

#### 4b. Summary handler test
- Test `GetBudgetSummary` with mixed-currency expenses → verify conversion to `display_currency`.
- Test with `preferred_quote_type = BLUE` → verify the correct quote is requested.

## Key Files

### Backend
- `core/internal/currency/provider.go` — `ExchangeRateProvider` interface, `CurrencyMarketplace`
- `core/internal/currency/stub_provider.go` — stub provider (remove `Supports`)
- `core/internal/currency/frankfurter_provider.go` — **NEW**: Frankfurter API provider
- `core/internal/currency/multi_provider.go` — **NEW**: chain-of-responsibility multi-provider
- `core/internal/currency/cache.go` — `InMemoryCache` (may need date-aware cache keys for historical)
- `core/internal/handler/expense_handler.go` — `GetBudgetSummary` (add conversion logic)
- `core/internal/server/server.go` — wire providers + marketplace into handlers
- `core/internal/config/config.go` — `ExchangeConfig` (already has Provider, APIKey, APIURL, Timeout)
- `core/internal/domain/persistible.go` — `PersistedUserPreferenceFromPersistence` (load user prefs)
- `core/internal/domain/models.go` — `QuoteType` constants, `Currency` constants

### Frontend
- `app/src/hooks/useBudgetSelection.ts` — replace client-side sums with summary API call
- `app/src/pages/BudgetDetail.tsx` — replace client-side sums with summary API call
- `app/src/lib/types.ts` — add `BudgetSummary` type
- `app/src/components/BudgetSummary.tsx` — no changes needed
- `app/src/lib/format.ts` — `formatCurrency` (formatting only, no conversion)
- `app/src/lib/usePreferences.ts` — `display_currency`, `preferred_quote_type` preferences

## External APIs

### Frankfurter (implement now)
- **Base URL**: `https://api.frankfurter.dev/v2`
- **Latest rates**: `GET /rates?base=USD&quotes=EUR,GBP`
- **Historical rates**: `GET /rates?base=USD&quotes=EUR&date=2024-01-15`
- **No API key required**
- **Response format**: `{ "base": "USD", "date": "2024-01-15", "rates": { "EUR": 0.92, "GBP": 0.79 } }`
- **Source**: European Central Bank + 98 other central banks

### DolarAPI (future, not implemented now)
- **Base URL**: `https://dolarapi.com/v1`
- **Current rates**: `GET /dolares` → array of `{moneda, casa, nombre, compra, venta, fechaActualizacion}`
- **Casa values**: `oficial`, `blue`, `bolsa` (MEP), `contadoconliqui` (CCL), `cripto`, `mayorista`, `tarjeta`
- **No API key required**
- **Historical ARS rates**: `https://argentinadatos.com/api/v1/cotizaciones/dolares/{casa}/{YYYY/MM/DD}` (separate API, same data source)
- **Mapping to QuoteType**: `OFFICIAL→oficial`, `BLUE→blue`, `MEP→bolsa`, `CCL→contadoconliqui`, `CRYPTO→cripto`
- **Only serves USD↔ARS** — other pairs need a different provider

## Implementation Status (as of 2026-09-15)

### Completed
- **Phase 1**: Provider architecture — `ExchangeRateProvider` interface, `FrankfurterProvider`, `MultiProvider`, `InMemoryCache` all implemented and wired in `server.go`.
- **Phase 2**: `GetBudgetSummary` in `expense_handler.go` now loads user preferences, decrypts amounts, and calls `marketplace.ConvertHistorical` for each expense. Expected expenses use budget `end_date` (or now), actual expenses use `expense_date`.
- **Phase 3**: Frontend calls `GET /budgets/:budget_id/summary` and displays converted totals. `BudgetSummary.tsx` receives backend-computed values. Display currency selector in nav bar wired to `PATCH /preferences`.
- **Phase 4 (partial)**: Selenium integration test `test_currency_conversion.py` written — creates budget with USD/EUR/ARS expenses, verifies original amounts in expense tables, verifies summary aggregation in display currency, verifies currency switch.

### Fixed: Frankfurter API v2 Response Format

**Resolved**: `frankfurterResponse` struct was expecting a JSON object with a `rates` map, but the API v2 returns a JSON array of `{date, base, quote, rate}` entries. Updated to `frankfurterRateEntry` struct and fixed `fetchRate`/`fetchRates` to decode `[]frankfurterRateEntry`.

### Test Results (2026-09-15)

Integration test `test_currency_conversion.py` — **PASSED** (88.78s):
- USD summary: Expected $1,622.10, Actual $1,510.44, Diff $111.67
- EUR summary: Expected 1.379,91 €, Actual 1.465,15 €, Diff -85,24 €
- ARS summary: Expected $ 2.360.230,00, Actual $ 1.573.192,50, Diff $ 787.037,50

All phases complete. Currency conversion is fully functional.

## Notes
- Use `docker compose` for all app build/run/logs commands.
- The `EXCHANGE_PROVIDER` env var controls which provider is wired (`"stub"` default, `"frankfurter"` for real rates).
- The `ExchangeConfig` already has `APIKey` and `APIURL` fields for future providers that need them.
- Frankfurter doesn't need an API key, so `ExchangeConfig.APIKey` can stay empty.
- The Frankfurter API v2 response format documented in the "External APIs" section above is **outdated** — the actual format is a JSON array of `{date, base, quote, rate}` entries (see "Fixed" section above).
