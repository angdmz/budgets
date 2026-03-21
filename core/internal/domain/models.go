package domain

import (
	"time"

	"github.com/google/uuid"
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
	switch c {
	case CurrencyUSD, CurrencyEUR, CurrencyGBP, CurrencyARS, CurrencyBRL,
		CurrencyMXN, CurrencyCLP, CurrencyCOP, CurrencyPEN, CurrencyUYU:
		return true
	}
	return false
}

// AuthProvider represents supported authentication providers
type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "GOOGLE"
	AuthProviderGitHub AuthProvider = "GITHUB"
	AuthProviderLocal  AuthProvider = "LOCAL"
)

type BaseModel struct {
	ID         int64      `json:"-"`
	ExternalID uuid.UUID  `json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	RevokedAt  *time.Time `json:"-"`
}

func (b *BaseModel) IsRevoked() bool {
	return b.RevokedAt != nil
}

// User represents an authenticated user in the system
type User struct {
	BaseModel
	ExternalProviderID string       `json:"-"`
	AuthProvider       AuthProvider `json:"provider"`
	Email              string       `json:"email"`
	DisplayName        string       `json:"display_name,omitempty"`
	AvatarURL          string       `json:"avatar_url,omitempty"`
}

type BudgetingGroup struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Participant represents a business-level participant in a group
// Multiple users can be associated with the same participant
type Participant struct {
	BaseModel
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	BudgetingGroupID int64  `json:"-"`
}

// UserParticipant represents the association between a User and a Participant
type UserParticipant struct {
	BaseModel
	UserID        int64  `json:"-"`
	ParticipantID int64  `json:"-"`
	Role          string `json:"role"`
	IsPrimary     bool   `json:"is_primary"`
}

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

// UserPreference stores per-user settings
type UserPreference struct {
	BaseModel
	UserID          int64    `json:"-"`
	Theme           Theme    `json:"theme"`
	Language        Language `json:"language"`
	DisplayCurrency Currency `json:"display_currency"`
}

type ExpenseCategory struct {
	BaseModel
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Color            string `json:"color,omitempty"`
	Icon             string `json:"icon,omitempty"`
	BudgetingGroupID int64  `json:"-"`
}

type Budget struct {
	BaseModel
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	BudgetingGroupID int64     `json:"-"`
}

// Money represents a monetary value with currency
type Money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency Currency        `json:"currency"`
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

type ExpectedExpense struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Amount      Money  `json:"amount"`
	BudgetID    int64  `json:"-"`
	CategoryID  *int64 `json:"-"`
}

type ActualExpense struct {
	BaseModel
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	ExpenseDate       time.Time `json:"expense_date"`
	Amount            Money     `json:"amount"`
	BudgetID          int64     `json:"-"`
	CategoryID        *int64    `json:"-"`
	ExpectedExpenseID *int64    `json:"-"`
}
