package domain

import (
	"github.com/shopspring/decimal"
)

// Currency represents supported currencies
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyARS Currency = "ARS"
	CurrencyBRL Currency = "BRL"
	CurrencyMXN Currency = "MXN"
	CurrencyCLP Currency = "CLP"
	CurrencyCOP Currency = "COP"
	CurrencyPEN Currency = "PEN"
	CurrencyUYU Currency = "UYU"
)

func (c Currency) IsValid() bool {
	for _, supported := range SupportedCurrencies() {
		if c == supported {
			return true
		}
	}
	return false
}

func SupportedCurrencies() []Currency {
	return []Currency{
		CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyARS, CurrencyBRL,
		CurrencyMXN, CurrencyCLP, CurrencyCOP, CurrencyPEN, CurrencyUYU,
	}
}

// QuoteType represents an exchange rate quote type
type QuoteType string

const (
	QuoteOfficial QuoteType = "OFFICIAL"
	QuoteBlue     QuoteType = "BLUE"
	QuoteMEP      QuoteType = "MEP"
	QuoteCCL      QuoteType = "CCL"
	QuoteCrypto   QuoteType = "CRYPTO"
)

func (q QuoteType) IsValid() bool {
	switch q {
	case QuoteOfficial, QuoteBlue, QuoteMEP, QuoteCCL, QuoteCrypto:
		return true
	}
	return false
}

func SupportedQuoteTypes() []QuoteType {
	return []QuoteType{QuoteOfficial, QuoteBlue, QuoteMEP, QuoteCCL, QuoteCrypto}
}

// AuthProvider represents supported authentication providers
type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "GOOGLE"
	AuthProviderGitHub AuthProvider = "GITHUB"
	AuthProviderLocal  AuthProvider = "LOCAL"
)

// ParticipantRole is the role a user plays inside a budgeting group.
type ParticipantRole string

const (
	ParticipantRoleOwner  ParticipantRole = "owner"
	ParticipantRoleMember ParticipantRole = "member"
)

// Theme represents UI theme options
type Theme string

const (
	ThemeLight Theme = "LIGHT"
	ThemeDim   Theme = "DIM"
	ThemeDark  Theme = "DARK"
)

func (t Theme) IsValid() bool {
	switch t {
	case ThemeLight, ThemeDim, ThemeDark:
		return true
	}
	return false
}

// Language represents supported languages
type Language string

const (
	LanguageEN Language = "EN"
	LanguageES Language = "ES"
)

func (l Language) IsValid() bool {
	switch l {
	case LanguageEN, LanguageES:
		return true
	}
	return false
}

// Money represents a monetary value with currency. It carries no JSON tags:
// the wire shape is representation.Money produced by Render methods.
type Money struct {
	Amount   decimal.Decimal
	Currency Currency
}

func NewMoney(amount decimal.Decimal, currency Currency) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

// Currency-specific money constructors
func NewUSDMoney(amount decimal.Decimal) Money {
	return NewMoney(amount, CurrencyUSD)
}

func NewARSMoney(amount decimal.Decimal) Money {
	return NewMoney(amount, CurrencyARS)
}

func NewEURMoney(amount decimal.Decimal) Money {
	return NewMoney(amount, CurrencyEUR)
}
