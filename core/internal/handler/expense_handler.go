package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/service"
)

type ExpenseHandler struct {
	expenseService service.ExpenseService
}

func NewExpenseHandler(expenseService service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

// CreateExpectedExpense godoc
// @Summary Create expected expense
// @Description Create a new expected expense for a budget
// @Tags expected-expenses
// @Accept json
// @Produce json
// @Param budget_id path string true "Budget ID (UUID)"
// @Param request body CreateExpectedExpenseRequest true "Expected expense request"
// @Success 201 {object} ExpectedExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{budget_id}/expected-expenses [post]
func (h *ExpenseHandler) CreateExpectedExpense(c *gin.Context) {
	budgetIDStr := c.Param("id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id"})
		return
	}

	var req CreateExpectedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	expense, err := h.expenseService.CreateExpected(c.Request.Context(), budgetID, req.Name, req.Description, amount, req.Amount.Currency, req.CategoryID, user.ExternalProviderID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ExpectedExpenseResponse{
		ID:          expense.ExternalID,
		Name:        expense.Name,
		Description: expense.Description,
		Amount:      MoneyResponse{Amount: expense.Amount.Amount.String(), Currency: string(expense.Amount.Currency)},
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	})
}

// GetExpectedExpenses godoc
// @Summary Get expected expenses
// @Description Get all expected expenses for a budget
// @Tags expected-expenses
// @Produce json
// @Param budget_id path string true "Budget ID (UUID)"
// @Success 200 {array} ExpectedExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{budget_id}/expected-expenses [get]
func (h *ExpenseHandler) GetExpectedExpenses(c *gin.Context) {
	budgetIDStr := c.Param("id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	expenses, err := h.expenseService.GetExpectedByBudgetID(c.Request.Context(), budgetID, user.ExternalProviderID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := make([]ExpectedExpenseResponse, len(expenses))
	for i, e := range expenses {
		response[i] = ExpectedExpenseResponse{
			ID:          e.ExternalID,
			Name:        e.Name,
			Description: e.Description,
			Amount:      MoneyResponse{Amount: e.Amount.Amount.String(), Currency: string(e.Amount.Currency)},
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateExpectedExpense godoc
// @Summary Update expected expense
// @Description Update an expected expense by ID
// @Tags expected-expenses
// @Accept json
// @Produce json
// @Param id path string true "Expected Expense ID (UUID)"
// @Param request body UpdateExpectedExpenseRequest true "Update request"
// @Success 200 {object} ExpectedExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /expected-expenses/{id} [put]
func (h *ExpenseHandler) UpdateExpectedExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id"})
		return
	}

	var req UpdateExpectedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	expense, err := h.expenseService.UpdateExpected(c.Request.Context(), id, req.Name, req.Description, amount, req.Amount.Currency, req.CategoryID, user.ExternalProviderID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, ExpectedExpenseResponse{
		ID:          expense.ExternalID,
		Name:        expense.Name,
		Description: expense.Description,
		Amount:      MoneyResponse{Amount: expense.Amount.Amount.String(), Currency: string(expense.Amount.Currency)},
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	})
}

// DeleteExpectedExpense godoc
// @Summary Delete expected expense
// @Description Soft delete an expected expense by ID
// @Tags expected-expenses
// @Param id path string true "Expected Expense ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /expected-expenses/{id} [delete]
func (h *ExpenseHandler) DeleteExpectedExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	if err := h.expenseService.DeleteExpected(c.Request.Context(), id, user.ExternalProviderID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateActualExpense godoc
// @Summary Create actual expense
// @Description Create a new actual expense for a budget
// @Tags actual-expenses
// @Accept json
// @Produce json
// @Param budget_id path string true "Budget ID (UUID)"
// @Param request body CreateActualExpenseRequest true "Actual expense request"
// @Success 201 {object} ActualExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{budget_id}/actual-expenses [post]
func (h *ExpenseHandler) CreateActualExpense(c *gin.Context) {
	budgetIDStr := c.Param("id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id"})
		return
	}

	var req CreateActualExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount"})
		return
	}

	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_expense_date"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	expense, err := h.expenseService.CreateActual(c.Request.Context(), budgetID, req.Name, req.Description, expenseDate, amount, req.Amount.Currency, req.CategoryID, req.ExpectedExpenseID, user.ExternalProviderID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ActualExpenseResponse{
		ID:          expense.ExternalID,
		Name:        expense.Name,
		Description: expense.Description,
		ExpenseDate: expense.ExpenseDate.Format("2006-01-02"),
		Amount:      MoneyResponse{Amount: expense.Amount.Amount.String(), Currency: string(expense.Amount.Currency)},
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	})
}

// GetActualExpenses godoc
// @Summary Get actual expenses
// @Description Get all actual expenses for a budget
// @Tags actual-expenses
// @Produce json
// @Param budget_id path string true "Budget ID (UUID)"
// @Success 200 {array} ActualExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{budget_id}/actual-expenses [get]
func (h *ExpenseHandler) GetActualExpenses(c *gin.Context) {
	budgetIDStr := c.Param("id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	expenses, err := h.expenseService.GetActualByBudgetID(c.Request.Context(), budgetID, user.ExternalProviderID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response := make([]ActualExpenseResponse, len(expenses))
	for i, e := range expenses {
		response[i] = ActualExpenseResponse{
			ID:          e.ExternalID,
			Name:        e.Name,
			Description: e.Description,
			ExpenseDate: e.ExpenseDate.Format("2006-01-02"),
			Amount:      MoneyResponse{Amount: e.Amount.Amount.String(), Currency: string(e.Amount.Currency)},
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateActualExpense godoc
// @Summary Update actual expense
// @Description Update an actual expense by ID
// @Tags actual-expenses
// @Accept json
// @Produce json
// @Param id path string true "Actual Expense ID (UUID)"
// @Param request body UpdateActualExpenseRequest true "Update request"
// @Success 200 {object} ActualExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /actual-expenses/{id} [put]
func (h *ExpenseHandler) UpdateActualExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id"})
		return
	}

	var req UpdateActualExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount"})
		return
	}

	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_expense_date"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	expense, err := h.expenseService.UpdateActual(c.Request.Context(), id, req.Name, req.Description, expenseDate, amount, req.Amount.Currency, req.CategoryID, req.ExpectedExpenseID, user.ExternalProviderID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, ActualExpenseResponse{
		ID:          expense.ExternalID,
		Name:        expense.Name,
		Description: expense.Description,
		ExpenseDate: expense.ExpenseDate.Format("2006-01-02"),
		Amount:      MoneyResponse{Amount: expense.Amount.Amount.String(), Currency: string(expense.Amount.Currency)},
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	})
}

// DeleteActualExpense godoc
// @Summary Delete actual expense
// @Description Soft delete an actual expense by ID
// @Tags actual-expenses
// @Param id path string true "Actual Expense ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /actual-expenses/{id} [delete]
func (h *ExpenseHandler) DeleteActualExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id"})
		return
	}

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	if err := h.expenseService.DeleteActual(c.Request.Context(), id, user.ExternalProviderID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleServiceError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found"})
		return
	}
	if errors.Is(err, domain.ErrForbidden) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
		return
	}
	SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
}
