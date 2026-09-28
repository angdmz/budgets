package handler

import (
	"github.com/google/uuid"

	"github.com/budgets/core/internal/representation"
)

// Response wire shapes live in the representation package; these aliases keep
// handler code and swagger annotations stable.
type (
	GroupResponse          = representation.Group
	CategoryResponse       = representation.Category
	BudgetResponse         = representation.Budget
	BudgetSummaryResponse  = representation.BudgetSummary
	MoneyResponse          = representation.Money
	ExpectedExpenseResponse = representation.ExpectedExpense
	ActualExpenseResponse  = representation.ActualExpense
	AuthCallbackResponse   = representation.AuthCallback
	PreferenceResponse     = representation.Preference
	ConvertCurrencyResponse = representation.ConvertCurrency
	ExchangeRateResponse   = representation.ExchangeRate
	InvitationResponse     = representation.Invitation
	InvitationDetailResponse = representation.InvitationDetail
	CurrencyResponse       = representation.Currency
	OnboardingResponse     = representation.Onboarding
	OnboardingStepResponse = representation.OnboardingStep
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

type CreateBudgetRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
}

type UpdateBudgetRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
}

type MoneyRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

type CreateExpectedExpenseRequest struct {
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	Amount      MoneyRequest `json:"amount" binding:"required"`
	CategoryID  uuid.UUID    `json:"category_id" binding:"required"`
}

type UpdateExpectedExpenseRequest struct {
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	Amount      MoneyRequest `json:"amount" binding:"required"`
	CategoryID  uuid.UUID    `json:"category_id" binding:"required"`
}

type CreateActualExpenseRequest struct {
	Name              string       `json:"name" binding:"required"`
	Description       string       `json:"description"`
	ExpenseDate       string       `json:"expense_date" binding:"required"`
	Amount            MoneyRequest `json:"amount" binding:"required"`
	CategoryID        uuid.UUID    `json:"category_id" binding:"required"`
	ExpectedExpenseID *uuid.UUID   `json:"expected_expense_id"`
}

type UpdateActualExpenseRequest struct {
	Name              string       `json:"name" binding:"required"`
	Description       string       `json:"description"`
	ExpenseDate       string       `json:"expense_date" binding:"required"`
	Amount            MoneyRequest `json:"amount" binding:"required"`
	CategoryID        uuid.UUID    `json:"category_id" binding:"required"`
	ExpectedExpenseID *uuid.UUID   `json:"expected_expense_id"`
}

type UpdatePreferenceRequest struct {
	Theme              string `json:"theme" binding:"required"`
	Language           string `json:"language" binding:"required"`
	DisplayCurrency    string `json:"display_currency" binding:"required"`
	PreferredQuoteType string `json:"preferred_quote_type"`
}

type PatchPreferenceRequest struct {
	Theme              *string `json:"theme"`
	Language           *string `json:"language"`
	DisplayCurrency    *string `json:"display_currency"`
	PreferredQuoteType *string `json:"preferred_quote_type"`
}

type ConvertCurrencyRequest struct {
	Amount       string `json:"amount" binding:"required"`
	FromCurrency string `json:"from_currency" binding:"required"`
	ToCurrency   string `json:"to_currency" binding:"required"`
	QuoteType    string `json:"quote_type"`
}

type CreateInvitationRequest struct {
	Role string `json:"role"`
}

type CompleteStepRequest struct {
	Data interface{} `json:"data"`
}

