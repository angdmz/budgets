package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestQuoteType_IsValid(t *testing.T) {
	valid := []QuoteType{QuoteOfficial, QuoteBlue, QuoteMEP, QuoteCCL, QuoteCrypto}
	for _, q := range valid {
		if !q.IsValid() {
			t.Errorf("expected %s to be valid", q)
		}
	}
	if QuoteType("INVALID").IsValid() {
		t.Error("expected INVALID to be invalid")
	}
	if QuoteType("").IsValid() {
		t.Error("expected empty to be invalid")
	}
}

func TestSupportedQuoteTypes(t *testing.T) {
	types := SupportedQuoteTypes()
	if len(types) != 5 {
		t.Errorf("expected 5 quote types, got %d", len(types))
	}
}

func TestNewPersistibleExchangeRate_Valid(t *testing.T) {
	observed := time.Now()
	er, err := NewPersistibleExchangeRate(
		CurrencyUSD, CurrencyARS,
		QuoteBlue,
		decimal.NewFromInt(1450),
		"stub",
		observed,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if er == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestNewPersistibleExchangeRate_InvalidFrom(t *testing.T) {
	_, err := NewPersistibleExchangeRate(
		Currency("XXX"), CurrencyARS,
		QuoteBlue,
		decimal.NewFromInt(1450),
		"stub",
		time.Now(),
	)
	if err == nil {
		t.Error("expected error for invalid from_currency")
	}
}

func TestNewPersistibleExchangeRate_InvalidTo(t *testing.T) {
	_, err := NewPersistibleExchangeRate(
		CurrencyUSD, Currency("XXX"),
		QuoteBlue,
		decimal.NewFromInt(1450),
		"stub",
		time.Now(),
	)
	if err == nil {
		t.Error("expected error for invalid to_currency")
	}
}

func TestNewPersistibleExchangeRate_InvalidQuote(t *testing.T) {
	_, err := NewPersistibleExchangeRate(
		CurrencyUSD, CurrencyARS,
		QuoteType("INVALID"),
		decimal.NewFromInt(1450),
		"stub",
		time.Now(),
	)
	if err == nil {
		t.Error("expected error for invalid quote type")
	}
}

func TestNewPersistibleExchangeRate_ZeroRate(t *testing.T) {
	_, err := NewPersistibleExchangeRate(
		CurrencyUSD, CurrencyARS,
		QuoteBlue,
		decimal.NewFromInt(0),
		"stub",
		time.Now(),
	)
	if err == nil {
		t.Error("expected error for zero rate")
	}
}

func TestNewPersistibleExchangeRate_NegativeRate(t *testing.T) {
	_, err := NewPersistibleExchangeRate(
		CurrencyUSD, CurrencyARS,
		QuoteBlue,
		decimal.NewFromInt(-1),
		"stub",
		time.Now(),
	)
	if err == nil {
		t.Error("expected error for negative rate")
	}
}

func TestNewPersistibleExchangeRate_EmptyProvider(t *testing.T) {
	_, err := NewPersistibleExchangeRate(
		CurrencyUSD, CurrencyARS,
		QuoteBlue,
		decimal.NewFromInt(1450),
		"",
		time.Now(),
	)
	if err == nil {
		t.Error("expected error for empty provider")
	}
}

func TestNewPersistibleExchangeRate_ZeroObservedAt(t *testing.T) {
	_, err := NewPersistibleExchangeRate(
		CurrencyUSD, CurrencyARS,
		QuoteBlue,
		decimal.NewFromInt(1450),
		"stub",
		time.Time{},
	)
	if err == nil {
		t.Error("expected error for zero observed_at")
	}
}

func TestNewPersistibleUserPreference_Valid(t *testing.T) {
	pref, err := NewPersistibleUserPreference(1, ThemeLight, LanguageEN, CurrencyUSD, QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pref == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewPersistibleUserPreference_InvalidQuoteType(t *testing.T) {
	_, err := NewPersistibleUserPreference(1, ThemeLight, LanguageEN, CurrencyUSD, QuoteType("INVALID"))
	if err == nil {
		t.Error("expected error for invalid quote type")
	}
}

func TestPersistedUserPreference_UpdatePreferredQuoteType(t *testing.T) {
	pref := &PersistedUserPreference{
		id:                1,
		externalID:        uuid.New(),
		userID:            1,
		theme:             ThemeLight,
		language:          LanguageEN,
		displayCurrency:   CurrencyUSD,
		preferredQuoteType: QuoteOfficial,
	}

	err := pref.UpdatePreferredQuoteType(QuoteBlue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pref.PreferredQuoteType() != QuoteBlue {
		t.Errorf("expected BLUE, got %s", pref.PreferredQuoteType())
	}
}

func TestPersistedUserPreference_UpdatePreferredQuoteType_Invalid(t *testing.T) {
	pref := &PersistedUserPreference{
		preferredQuoteType: QuoteOfficial,
	}

	err := pref.UpdatePreferredQuoteType(QuoteType("INVALID"))
	if err == nil {
		t.Error("expected error for invalid quote type")
	}
	if pref.PreferredQuoteType() != QuoteOfficial {
		t.Error("expected quote type to remain unchanged on error")
	}
}
