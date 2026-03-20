package handler

import (
	"time"

	"github.com/google/uuid"
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

type GroupResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

type CategoryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

type BudgetResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MoneyRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

type MoneyResponse struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type CreateExpectedExpenseRequest struct {
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	Amount      MoneyRequest `json:"amount" binding:"required"`
	CategoryID  *uuid.UUID   `json:"category_id"`
}

type UpdateExpectedExpenseRequest struct {
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	Amount      MoneyRequest `json:"amount" binding:"required"`
	CategoryID  *uuid.UUID   `json:"category_id"`
}

type ExpectedExpenseResponse struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Amount      MoneyResponse `json:"amount"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type CreateActualExpenseRequest struct {
	Name              string       `json:"name" binding:"required"`
	Description       string       `json:"description"`
	ExpenseDate       string       `json:"expense_date" binding:"required"`
	Amount            MoneyRequest `json:"amount" binding:"required"`
	CategoryID        *uuid.UUID   `json:"category_id"`
	ExpectedExpenseID *uuid.UUID   `json:"expected_expense_id"`
}

type UpdateActualExpenseRequest struct {
	Name              string       `json:"name" binding:"required"`
	Description       string       `json:"description"`
	ExpenseDate       string       `json:"expense_date" binding:"required"`
	Amount            MoneyRequest `json:"amount" binding:"required"`
	CategoryID        *uuid.UUID   `json:"category_id"`
	ExpectedExpenseID *uuid.UUID   `json:"expected_expense_id"`
}

type ActualExpenseResponse struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	ExpenseDate string        `json:"expense_date"`
	Amount      MoneyResponse `json:"amount"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type AuthCallbackResponse struct {
	Token string `json:"token"`
}
