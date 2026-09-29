package domain

import (
	"testing"

	"github.com/google/uuid"
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

func TestNewPersistibleUserPreference_Valid(t *testing.T) {
	pref, err := NewPersistibleUserPreference(&PersistedUser{id: 1}, ThemeLight, LanguageEN, CurrencyUSD, QuoteOfficial)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pref == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewPersistibleUserPreference_InvalidQuoteType(t *testing.T) {
	_, err := NewPersistibleUserPreference(&PersistedUser{id: 1}, ThemeLight, LanguageEN, CurrencyUSD, QuoteType("INVALID"))
	if err == nil {
		t.Error("expected error for invalid quote type")
	}
}

func TestNewPersistibleUserPreference_NilUser(t *testing.T) {
	_, err := NewPersistibleUserPreference(nil, ThemeLight, LanguageEN, CurrencyUSD, QuoteOfficial)
	if err == nil {
		t.Error("expected error for nil user")
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
	if pref.Presentation().preferredQuote != QuoteBlue {
		t.Errorf("expected BLUE, got %s", pref.Presentation().preferredQuote)
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
	if pref.Presentation().preferredQuote != QuoteOfficial {
		t.Error("expected quote type to remain unchanged on error")
	}
}
