package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PersistibleExchangeRate is a new exchange rate observation to be persisted.
type PersistibleExchangeRate struct {
	fromCurrency Currency
	toCurrency   Currency
	quote        QuoteType
	rate         decimal.Decimal
	provider     string
	observedAt   time.Time
}

func NewPersistibleExchangeRate(
	from, to Currency,
	quote QuoteType,
	rate decimal.Decimal,
	provider string,
	observedAt time.Time,
) (*PersistibleExchangeRate, error) {
	if !from.IsValid() {
		return nil, fmt.Errorf("%w: invalid from_currency", ErrValidation)
	}
	if !to.IsValid() {
		return nil, fmt.Errorf("%w: invalid to_currency", ErrValidation)
	}
	if !quote.IsValid() {
		return nil, fmt.Errorf("%w: invalid quote type", ErrValidation)
	}
	if rate.IsZero() || rate.IsNegative() {
		return nil, fmt.Errorf("%w: rate must be positive", ErrValidation)
	}
	if provider == "" {
		return nil, fmt.Errorf("%w: provider cannot be empty", ErrValidation)
	}
	if observedAt.IsZero() {
		return nil, fmt.Errorf("%w: observed_at cannot be zero", ErrValidation)
	}
	return &PersistibleExchangeRate{
		fromCurrency: from,
		toCurrency:   to,
		quote:        quote,
		rate:         rate,
		provider:     provider,
		observedAt:   observedAt,
	}, nil
}

func (e *PersistibleExchangeRate) PersistTo(ctx context.Context, p Persister) (*PersistedExchangeRate, error) {
	var id int64
	var externalID uuid.UUID
	var createdAt, updatedAt time.Time

	err := p.QueryRow(
		ctx,
		[]any{&id, &externalID, &createdAt, &updatedAt},
		`INSERT INTO exchange_rates (from_currency, to_currency, quote, rate, provider, observed_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (from_currency, to_currency, quote) WHERE revoked_at IS NULL
		 DO UPDATE SET rate = EXCLUDED.rate, provider = EXCLUDED.provider, observed_at = EXCLUDED.observed_at, updated_at = CURRENT_TIMESTAMP
		 RETURNING id, external_id, created_at, updated_at`,
		e.fromCurrency, e.toCurrency, e.quote, e.rate, e.provider, e.observedAt,
	)
	if err != nil {
		return nil, err
	}

	return &PersistedExchangeRate{
		id:           id,
		externalID:   externalID,
		fromCurrency: e.fromCurrency,
		toCurrency:   e.toCurrency,
		quote:        e.quote,
		rate:         e.rate,
		provider:     e.provider,
		observedAt:   e.observedAt,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

// PersistedExchangeRate is an exchange rate loaded from the database.
type PersistedExchangeRate struct {
	id           int64
	externalID   uuid.UUID
	fromCurrency Currency
	toCurrency   Currency
	quote        QuoteType
	rate         decimal.Decimal
	provider     string
	observedAt   time.Time
	createdAt    time.Time
	updatedAt    time.Time
}

func PersistedExchangeRateFromPersistence(
	ctx context.Context,
	from, to Currency,
	quote QuoteType,
	p Persister,
) (*PersistedExchangeRate, error) {
	var e PersistedExchangeRate
	err := p.QueryRow(
		ctx,
		[]any{&e.id, &e.externalID, &e.rate, &e.provider, &e.observedAt, &e.createdAt, &e.updatedAt},
		`SELECT id, external_id, rate, provider, observed_at, created_at, updated_at
		 FROM exchange_rates
		 WHERE from_currency = $1 AND to_currency = $2 AND quote = $3 AND revoked_at IS NULL
		 ORDER BY observed_at DESC
		 LIMIT 1`,
		from, to, quote,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: exchange rate not found", ErrNotFound)
	}
	e.fromCurrency = from
	e.toCurrency = to
	e.quote = quote
	return &e, nil
}

func (e *PersistedExchangeRate) ID() int64              { return e.id }
func (e *PersistedExchangeRate) ExternalID() uuid.UUID  { return e.externalID }
func (e *PersistedExchangeRate) FromCurrency() Currency { return e.fromCurrency }
func (e *PersistedExchangeRate) ToCurrency() Currency   { return e.toCurrency }
func (e *PersistedExchangeRate) Quote() QuoteType       { return e.quote }
func (e *PersistedExchangeRate) Rate() decimal.Decimal   { return e.rate }
func (e *PersistedExchangeRate) Provider() string        { return e.provider }
func (e *PersistedExchangeRate) ObservedAt() time.Time   { return e.observedAt }
func (e *PersistedExchangeRate) CreatedAt() time.Time    { return e.createdAt }
func (e *PersistedExchangeRate) UpdatedAt() time.Time    { return e.updatedAt }
