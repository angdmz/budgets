package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/middleware"
)

type ExpenseHandler struct {
	pool      *pgxpool.Pool
	encryptor *encryption.Encryptor
}

func NewExpenseHandler(pool *pgxpool.Pool, encryptor *encryption.Encryptor) *ExpenseHandler {
	return &ExpenseHandler{pool: pool, encryptor: encryptor}
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
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	budgetIDStr := c.Param("budget_id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id", Message: "Invalid UUID format"})
		return
	}

	var req CreateExpectedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Invalid amount format"})
		return
	}

	var response ExpectedExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			return err
		}

		money := encryption.NewMoney(amount, req.Amount.Currency)
		encryptedAmount, err := h.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		persistibleExpense, err := domain.NewPersistibleExpectedExpense(req.Name, req.Description, encryptedAmount, budgetID, req.CategoryID)
		if err != nil {
			return err
		}

		persistedExpense, err := persistibleExpense.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		decryptedMoney, err := h.encryptor.DecryptMoney(persistedExpense.EncryptedAmount())
		if err != nil {
			return err
		}

		response = ExpectedExpenseResponse{
			ID:          persistedExpense.ExternalID(),
			Name:        persistedExpense.Name(),
			Description: persistedExpense.Description(),
			Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
			CategoryID:  persistedExpense.CategoryExternalID(),
			CreatedAt:   persistedExpense.CreatedAt(),
			UpdatedAt:   persistedExpense.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetExpectedExpense godoc
// @Summary Get a specific expected expense
// @Description Get an expected expense by ID
// @Tags expected-expenses
// @Produce json
// @Param id path string true "Expected Expense ID (UUID)"
// @Success 200 {object} ExpectedExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /expected-expenses/{id} [get]
func (h *ExpenseHandler) GetExpectedExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response ExpectedExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeExpenseAccess(ctx, p, id); err != nil {
			return err
		}

		expense, err := domain.PersistedExpectedExpenseFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		decryptedMoney, err := h.encryptor.DecryptMoney(expense.EncryptedAmount())
		if err != nil {
			return err
		}

		response = ExpectedExpenseResponse{
			ID:          expense.ExternalID(),
			Name:        expense.Name(),
			Description: expense.Description(),
			Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
			CategoryID:  expense.CategoryExternalID(),
			CreatedAt:   expense.CreatedAt(),
			UpdatedAt:   expense.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
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
	budgetIDStr := c.Param("budget_id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response []ExpectedExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			return err
		}

		expenses, err := domain.PersistedExpectedExpensesForBudget(ctx, budgetID, p)
		if err != nil {
			return err
		}

		response = make([]ExpectedExpenseResponse, len(expenses))
		for i, e := range expenses {
			decryptedMoney, err := h.encryptor.DecryptMoney(e.EncryptedAmount())
			if err != nil {
				return err
			}

			response[i] = ExpectedExpenseResponse{
				ID:          e.ExternalID(),
				Name:        e.Name(),
				Description: e.Description(),
				Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
				CategoryID:  e.CategoryExternalID(),
				CreatedAt:   e.CreatedAt(),
				UpdatedAt:   e.UpdatedAt(),
			}
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req UpdateExpectedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Invalid amount format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response ExpectedExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeExpenseAccess(ctx, p, id); err != nil {
			return err
		}

		expense, err := domain.PersistedExpectedExpenseFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		money := encryption.NewMoney(amount, req.Amount.Currency)
		encryptedAmount, err := h.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		expense.UpdateName(req.Name)
		expense.UpdateDescription(req.Description)
		expense.UpdateEncryptedAmount(encryptedAmount)
		expense.UpdateCategoryExternalID(req.CategoryID)

		if err := expense.UpdateIn(ctx, p); err != nil {
			return err
		}

		decryptedMoney, err := h.encryptor.DecryptMoney(expense.EncryptedAmount())
		if err != nil {
			return err
		}

		response = ExpectedExpenseResponse{
			ID:          expense.ExternalID(),
			Name:        expense.Name(),
			Description: expense.Description(),
			Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
			CategoryID:  expense.CategoryExternalID(),
			CreatedAt:   expense.CreatedAt(),
			UpdatedAt:   expense.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeExpenseAccess(ctx, p, id); err != nil {
			return err
		}

		expense, err := domain.PersistedExpectedExpenseFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		return expense.DeleteFrom(ctx, p)
	})

	if err != nil {
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
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	budgetIDStr := c.Param("budget_id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id", Message: "Invalid UUID format"})
		return
	}

	var req CreateActualExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Invalid amount format"})
		return
	}

	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_expense_date", Message: "Date must be in YYYY-MM-DD format"})
		return
	}

	var response ActualExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			return err
		}

		money := encryption.NewMoney(amount, req.Amount.Currency)
		encryptedAmount, err := h.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		persistibleExpense, err := domain.NewPersistibleActualExpense(req.Name, req.Description, expenseDate, encryptedAmount, budgetID, req.CategoryID, req.ExpectedExpenseID)
		if err != nil {
			return err
		}

		persistedExpense, err := persistibleExpense.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		decryptedMoney, err := h.encryptor.DecryptMoney(persistedExpense.EncryptedAmount())
		if err != nil {
			return err
		}

		response = ActualExpenseResponse{
			ID:          persistedExpense.ExternalID(),
			Name:        persistedExpense.Name(),
			Description: persistedExpense.Description(),
			ExpenseDate: persistedExpense.ExpenseDate().Format("2006-01-02"),
			Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
			CreatedAt:   persistedExpense.CreatedAt(),
			UpdatedAt:   persistedExpense.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetActualExpense godoc
// @Summary Get a specific actual expense
// @Description Get an actual expense by ID
// @Tags actual-expenses
// @Produce json
// @Param id path string true "Actual Expense ID (UUID)"
// @Success 200 {object} ActualExpenseResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /actual-expenses/{id} [get]
func (h *ExpenseHandler) GetActualExpense(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response ActualExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeExpenseAccess(ctx, p, id); err != nil {
			return err
		}

		expense, err := domain.PersistedActualExpenseFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		decryptedMoney, err := h.encryptor.DecryptMoney(expense.EncryptedAmount())
		if err != nil {
			return err
		}

		response = ActualExpenseResponse{
			ID:          expense.ExternalID(),
			Name:        expense.Name(),
			Description: expense.Description(),
			ExpenseDate: expense.ExpenseDate().Format("2006-01-02"),
			Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
			CreatedAt:   expense.CreatedAt(),
			UpdatedAt:   expense.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
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
	budgetIDStr := c.Param("budget_id")
	budgetID, err := uuid.Parse(budgetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_budget_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response []ActualExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			return err
		}

		expenses, err := domain.PersistedActualExpensesForBudget(ctx, budgetID, p)
		if err != nil {
			return err
		}

		response = make([]ActualExpenseResponse, len(expenses))
		for i, e := range expenses {
			decryptedMoney, err := h.encryptor.DecryptMoney(e.EncryptedAmount())
			if err != nil {
				return err
			}

			response[i] = ActualExpenseResponse{
				ID:          e.ExternalID(),
				Name:        e.Name(),
				Description: e.Description(),
				ExpenseDate: e.ExpenseDate().Format("2006-01-02"),
				Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
				CreatedAt:   e.CreatedAt(),
				UpdatedAt:   e.UpdatedAt(),
			}
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req UpdateActualExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, err := decimal.NewFromString(req.Amount.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_amount", Message: "Invalid amount format"})
		return
	}

	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_expense_date", Message: "Date must be in YYYY-MM-DD format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response ActualExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeExpenseAccess(ctx, p, id); err != nil {
			return err
		}

		expense, err := domain.PersistedActualExpenseFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		money := encryption.NewMoney(amount, req.Amount.Currency)
		encryptedAmount, err := h.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		expense.UpdateName(req.Name)
		expense.UpdateDescription(req.Description)
		expense.UpdateExpenseDate(expenseDate)
		expense.UpdateEncryptedAmount(encryptedAmount)

		if err := expense.UpdateIn(ctx, p); err != nil {
			return err
		}

		decryptedMoney, err := h.encryptor.DecryptMoney(expense.EncryptedAmount())
		if err != nil {
			return err
		}

		response = ActualExpenseResponse{
			ID:          expense.ExternalID(),
			Name:        expense.Name(),
			Description: expense.Description(),
			ExpenseDate: expense.ExpenseDate().Format("2006-01-02"),
			Amount:      MoneyResponse{Amount: decryptedMoney.Amount.String(), Currency: decryptedMoney.Currency},
			CreatedAt:   expense.CreatedAt(),
			UpdatedAt:   expense.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeExpenseAccess(ctx, p, id); err != nil {
			return err
		}

		expense, err := domain.PersistedActualExpenseFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		return expense.DeleteFrom(ctx, p)
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleServiceError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Resource not found"})
		return
	}
	if errors.Is(err, domain.ErrForbidden) {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to perform this action"})
		return
	}
	SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
}
