// Package representation holds pure wire shapes for API responses.
// Types in this package carry no domain logic and no domain dependencies:
// they are the target of domain.Rendered[T] produced by Render() methods
// on Persisted* domain objects.
package representation

import (
	"time"

	"github.com/google/uuid"
)

// User is the wire shape for an authenticated user.
type User struct {
	ID          uuid.UUID `json:"id"`
	Provider    string    `json:"provider"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Money is the wire shape for a monetary amount.
type Money struct {
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
	Converted bool   `json:"converted,omitempty"`
}

// Group is the wire shape for a budgeting group.
type Group struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Category is the wire shape for an expense category.
type Category struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Budget is the wire shape for a budget.
type Budget struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BudgetSummary is the wire shape for a budget's expense totals.
type BudgetSummary struct {
	BudgetID      uuid.UUID `json:"budget_id"`
	ExpectedTotal Money     `json:"expected_total"`
	ActualTotal   Money     `json:"actual_total"`
	Difference    Money     `json:"difference"`
}

// ExpectedExpense is the wire shape for an expected expense.
type ExpectedExpense struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	Amount          Money     `json:"amount"`
	ConvertedAmount *Money    `json:"converted_amount,omitempty"`
	CategoryID      uuid.UUID `json:"category_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ActualExpense is the wire shape for an actual expense.
type ActualExpense struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	ExpenseDate     string    `json:"expense_date"`
	Amount          Money     `json:"amount"`
	ConvertedAmount *Money    `json:"converted_amount,omitempty"`
	CategoryID      uuid.UUID `json:"category_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AuthCallback is the wire shape for an auth callback token response.
type AuthCallback struct {
	Token string `json:"token"`
}

// Preference is the wire shape for user preferences.
type Preference struct {
	Theme              string `json:"theme"`
	Language           string `json:"language"`
	DisplayCurrency    string `json:"display_currency"`
	PreferredQuoteType string `json:"preferred_quote_type"`
}

// ConvertCurrency is the wire shape for a currency conversion result.
type ConvertCurrency struct {
	OriginalAmount  Money  `json:"original_amount"`
	ConvertedAmount Money  `json:"converted_amount"`
	ExchangeRate    string `json:"exchange_rate"`
	Provider        string `json:"provider"`
	QuoteType       string `json:"quote_type,omitempty"`
}

// ExchangeRate is the wire shape for a quoted exchange rate.
type ExchangeRate struct {
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	QuoteType    string `json:"quote_type,omitempty"`
	Rate         string `json:"rate"`
	Provider     string `json:"provider"`
}

// Invitation is the wire shape for a group invitation.
type Invitation struct {
	ID          uuid.UUID  `json:"id"`
	Token       string     `json:"token"`
	GroupID     uuid.UUID  `json:"group_id,omitempty"`
	GroupName   string     `json:"group_name"`
	InviterName string     `json:"inviter_name"`
	Status      string     `json:"status"`
	Role        string     `json:"role"`
	ExpiresAt   time.Time  `json:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// InvitationDetail is the wire shape for a public invitation lookup.
type InvitationDetail struct {
	GroupName   string    `json:"group_name"`
	InviterName string    `json:"inviter_name"`
	Status      string    `json:"status"`
	Role        string    `json:"role"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Currency is the wire shape for a supported currency.
type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Onboarding is the wire shape for a user's onboarding state.
type Onboarding struct {
	ID          uuid.UUID        `json:"id"`
	Status      string           `json:"status"`
	CurrentStep string           `json:"current_step"`
	Steps       []OnboardingStep `json:"steps"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// OnboardingStep is the wire shape for a single onboarding step record.
type OnboardingStep struct {
	Step      string      `json:"step"`
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
