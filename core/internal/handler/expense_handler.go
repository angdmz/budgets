package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/budgets/core/internal/currency"
	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/middleware"
)

type ExpenseHandler struct {
	pool        *pgxpool.Pool
	encryptor   *encryption.Encryptor
	marketplace *currency.CurrencyMarketplace
}

func NewExpenseHandler(pool *pgxpool.Pool, encryptor *encryption.Encryptor, marketplace *currency.CurrencyMarketplace) *ExpenseHandler {
	return &ExpenseHandler{pool: pool, encryptor: encryptor, marketplace: marketplace}
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

	amount, currency, ok := validateMoneyRequest(c, req.Amount)
	if !ok {
		return
	}

	var response ExpectedExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			return err
		}

		money := encryption.NewMoney(amount, string(currency))
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

		response = toExpectedExpenseResponse(persistedExpense, decryptedMoney, nil)
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

		response = toExpectedExpenseResponse(expense, decryptedMoney, nil)
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

	log.Printf("[DEBUG] GetExpectedExpenses: budgetID=%s userID=%d", budgetID, user.ID)

	var expenses []domain.PersistedExpectedExpense
	displayCurrency := domain.CurrencyUSD
	preferredQuote := domain.QuoteOfficial
	var expectedDate time.Time

	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		log.Printf("[DEBUG] GetExpectedExpenses: inside transaction, authorizing budget access")
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			log.Printf("[DEBUG] GetExpectedExpenses: AuthorizeBudgetAccess failed: %v", err)
			return err
		}

		var eErr error
		expenses, eErr = domain.PersistedExpectedExpensesForBudget(ctx, budgetID, p)
		if eErr != nil {
			log.Printf("[DEBUG] GetExpectedExpenses: PersistedExpectedExpensesForBudget failed: %v", eErr)
			return eErr
		}
		log.Printf("[DEBUG] GetExpectedExpenses: loaded %d expenses", len(expenses))

		pref, pErr := domain.PersistedUserPreferenceFromPersistence(ctx, user.ID, p)
		if pErr == nil {
			displayCurrency = pref.DisplayCurrency()
			preferredQuote = pref.PreferredQuoteType()
		}
		log.Printf("[DEBUG] GetExpectedExpenses: displayCurrency=%s preferredQuote=%s prefErr=%v", displayCurrency, preferredQuote, pErr)

		budget, bErr := domain.PersistedBudgetFromPersistence(ctx, budgetID, p)
		if bErr == nil {
			expectedDate = budget.EndDate()
			if expectedDate.After(time.Now()) {
				expectedDate = time.Now()
			}
		}
		log.Printf("[DEBUG] GetExpectedExpenses: expectedDate=%v budgetErr=%v", expectedDate, bErr)

		log.Printf("[DEBUG] GetExpectedExpenses: returning nil from transaction fn")
		return nil
	})

	if err != nil {
		log.Printf("[DEBUG] GetExpectedExpenses: WithPersister returned error: %v", err)
		handleServiceError(c, err)
		return
	}
	log.Printf("[DEBUG] GetExpectedExpenses: transaction committed OK, converting %d expenses", len(expenses))

	response := make([]ExpectedExpenseResponse, len(expenses))
	for i, e := range expenses {
		decryptedMoney, dErr := h.encryptor.DecryptMoney(e.EncryptedAmount())
		if dErr != nil {
			log.Printf("[DEBUG] GetExpectedExpenses: DecryptMoney failed for expense %d: %v", i, dErr)
			handleServiceError(c, dErr)
			return
		}

		var convertedAmount *MoneyResponse
		origCurrency := domain.Currency(decryptedMoney.Currency)
		if origCurrency != displayCurrency && h.marketplace != nil && !expectedDate.IsZero() {
			log.Printf("[DEBUG] GetExpectedExpenses: converting expense %d: %s %s -> %s (date=%v)", i, decryptedMoney.Amount.String(), origCurrency, displayCurrency, expectedDate)
			money := domain.NewMoney(decryptedMoney.Amount, origCurrency)
			converted, cErr := h.marketplace.ConvertHistorical(c.Request.Context(), money, displayCurrency, preferredQuote, expectedDate)
			if cErr == nil {
				convertedAmount = &MoneyResponse{
					Amount:    converted.Amount.String(),
					Currency:  string(converted.Currency),
					Converted: true,
				}
				log.Printf("[DEBUG] GetExpectedExpenses: conversion OK: %s %s -> %s %s", decryptedMoney.Amount.String(), origCurrency, converted.Amount.String(), converted.Currency)
			} else {
				log.Printf("[DEBUG] GetExpectedExpenses: conversion FAILED for expense %d: %v", i, cErr)
			}
		}

		response[i] = toExpectedExpenseResponse(&e, decryptedMoney, convertedAmount)
	}

	log.Printf("[DEBUG] GetExpectedExpenses: sending response with %d items", len(response))
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

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var req UpdateExpectedExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, currency, ok := validateMoneyRequest(c, req.Amount)
	if !ok {
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

		money := encryption.NewMoney(amount, string(currency))
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

		response = toExpectedExpenseResponse(expense, decryptedMoney, nil)
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

	amount, currency, ok := validateMoneyRequest(c, req.Amount)
	if !ok {
		return
	}

	expenseDate, ok := parseExpenseDate(c, req.ExpenseDate)
	if !ok {
		return
	}

	var response ActualExpenseResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			return err
		}

		money := encryption.NewMoney(amount, string(currency))
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

		response = toActualExpenseResponse(persistedExpense, decryptedMoney, nil)
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

		response = toActualExpenseResponse(expense, decryptedMoney, nil)
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

	log.Printf("[DEBUG] GetActualExpenses: budgetID=%s userID=%d", budgetID, user.ID)

	var expenses []domain.PersistedActualExpense
	displayCurrency := domain.CurrencyUSD
	preferredQuote := domain.QuoteOfficial

	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		log.Printf("[DEBUG] GetActualExpenses: inside transaction, authorizing budget access")
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			log.Printf("[DEBUG] GetActualExpenses: AuthorizeBudgetAccess failed: %v", err)
			return err
		}

		var eErr error
		expenses, eErr = domain.PersistedActualExpensesForBudget(ctx, budgetID, p)
		if eErr != nil {
			log.Printf("[DEBUG] GetActualExpenses: PersistedActualExpensesForBudget failed: %v", eErr)
			return eErr
		}
		log.Printf("[DEBUG] GetActualExpenses: loaded %d expenses", len(expenses))

		pref, pErr := domain.PersistedUserPreferenceFromPersistence(ctx, user.ID, p)
		if pErr == nil {
			displayCurrency = pref.DisplayCurrency()
			preferredQuote = pref.PreferredQuoteType()
		}
		log.Printf("[DEBUG] GetActualExpenses: displayCurrency=%s preferredQuote=%s prefErr=%v", displayCurrency, preferredQuote, pErr)

		log.Printf("[DEBUG] GetActualExpenses: returning nil from transaction fn")
		return nil
	})

	if err != nil {
		log.Printf("[DEBUG] GetActualExpenses: WithPersister returned error: %v", err)
		handleServiceError(c, err)
		return
	}
	log.Printf("[DEBUG] GetActualExpenses: transaction committed OK, converting %d expenses", len(expenses))

	response := make([]ActualExpenseResponse, len(expenses))
	for i, e := range expenses {
		decryptedMoney, dErr := h.encryptor.DecryptMoney(e.EncryptedAmount())
		if dErr != nil {
			log.Printf("[DEBUG] GetActualExpenses: DecryptMoney failed for expense %d: %v", i, dErr)
			handleServiceError(c, dErr)
			return
		}

		var convertedAmount *MoneyResponse
		origCurrency := domain.Currency(decryptedMoney.Currency)
		if origCurrency != displayCurrency && h.marketplace != nil {
			log.Printf("[DEBUG] GetActualExpenses: converting expense %d: %s %s -> %s (date=%v)", i, decryptedMoney.Amount.String(), origCurrency, displayCurrency, e.ExpenseDate())
			money := domain.NewMoney(decryptedMoney.Amount, origCurrency)
			converted, cErr := h.marketplace.ConvertHistorical(c.Request.Context(), money, displayCurrency, preferredQuote, e.ExpenseDate())
			if cErr == nil {
				convertedAmount = &MoneyResponse{
					Amount:    converted.Amount.String(),
					Currency:  string(converted.Currency),
					Converted: true,
				}
				log.Printf("[DEBUG] GetActualExpenses: conversion OK: %s %s -> %s %s", decryptedMoney.Amount.String(), origCurrency, converted.Amount.String(), converted.Currency)
			} else {
				log.Printf("[DEBUG] GetActualExpenses: conversion FAILED for expense %d: %v", i, cErr)
			}
		}

		response[i] = toActualExpenseResponse(&e, decryptedMoney, convertedAmount)
	}

	log.Printf("[DEBUG] GetActualExpenses: sending response with %d items", len(response))
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

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var req UpdateActualExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	amount, currency, ok := validateMoneyRequest(c, req.Amount)
	if !ok {
		return
	}

	expenseDate, ok := parseExpenseDate(c, req.ExpenseDate)
	if !ok {
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

		money := encryption.NewMoney(amount, string(currency))
		encryptedAmount, err := h.encryptor.EncryptMoney(money)
		if err != nil {
			return err
		}

		expense.UpdateName(req.Name)
		expense.UpdateDescription(req.Description)
		expense.UpdateExpenseDate(expenseDate)
		expense.UpdateEncryptedAmount(encryptedAmount)
		expense.UpdateCategoryExternalID(req.CategoryID)

		if err := expense.UpdateIn(ctx, p); err != nil {
			return err
		}

		decryptedMoney, err := h.encryptor.DecryptMoney(expense.EncryptedAmount())
		if err != nil {
			return err
		}

		response = toActualExpenseResponse(expense, decryptedMoney, nil)
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
	if errors.Is(err, domain.ErrValidation) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}
	SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
}

// GetBudgetSummary godoc
// @Summary Get budget summary
// @Description Get expected total, actual total, and difference for a budget
// @Tags budgets
// @Produce json
// @Param budget_id path string true "Budget ID (UUID)"
// @Success 200 {object} BudgetSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{budget_id}/summary [get]
func (h *ExpenseHandler) GetBudgetSummary(c *gin.Context) {
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

	log.Printf("[DEBUG] GetBudgetSummary: budgetID=%s userID=%d", budgetID, user.ID)

	displayCurrency := domain.CurrencyUSD
	preferredQuote := domain.QuoteOfficial
	var expectedDate time.Time
	var expectedExpenses []domain.PersistedExpectedExpense
	var actualExpenses []domain.PersistedActualExpense

	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		log.Printf("[DEBUG] GetBudgetSummary: inside transaction, authorizing budget access")
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, budgetID); err != nil {
			log.Printf("[DEBUG] GetBudgetSummary: AuthorizeBudgetAccess failed: %v", err)
			return err
		}

		pref, pErr := domain.PersistedUserPreferenceFromPersistence(ctx, user.ID, p)
		if pErr == nil {
			displayCurrency = pref.DisplayCurrency()
			preferredQuote = pref.PreferredQuoteType()
		}
		log.Printf("[DEBUG] GetBudgetSummary: displayCurrency=%s preferredQuote=%s prefErr=%v", displayCurrency, preferredQuote, pErr)

		budget, bErr := domain.PersistedBudgetFromPersistence(ctx, budgetID, p)
		if bErr != nil {
			log.Printf("[DEBUG] GetBudgetSummary: PersistedBudgetFromPersistence failed: %v", bErr)
			return bErr
		}

		expectedDate = budget.EndDate()
		if expectedDate.After(time.Now()) {
			expectedDate = time.Now()
		}
		log.Printf("[DEBUG] GetBudgetSummary: expectedDate=%v", expectedDate)

		var eeErr error
		expectedExpenses, eeErr = domain.PersistedExpectedExpensesForBudget(ctx, budgetID, p)
		if eeErr != nil {
			log.Printf("[DEBUG] GetBudgetSummary: PersistedExpectedExpensesForBudget failed: %v", eeErr)
			return eeErr
		}
		log.Printf("[DEBUG] GetBudgetSummary: loaded %d expected expenses", len(expectedExpenses))

		var aeErr error
		actualExpenses, aeErr = domain.PersistedActualExpensesForBudget(ctx, budgetID, p)
		if aeErr != nil {
			log.Printf("[DEBUG] GetBudgetSummary: PersistedActualExpensesForBudget failed: %v", aeErr)
			return aeErr
		}
		log.Printf("[DEBUG] GetBudgetSummary: loaded %d actual expenses", len(actualExpenses))

		log.Printf("[DEBUG] GetBudgetSummary: returning nil from transaction fn")
		return nil
	})

	if err != nil {
		log.Printf("[DEBUG] GetBudgetSummary: WithPersister returned error: %v", err)
		handleServiceError(c, err)
		return
	}
	log.Printf("[DEBUG] GetBudgetSummary: transaction committed OK, converting expenses")

	expectedTotal := decimal.Zero
	for _, e := range expectedExpenses {
		decrypted, dErr := h.encryptor.DecryptMoney(e.EncryptedAmount())
		if dErr != nil {
			handleServiceError(c, dErr)
			return
		}
		money := domain.NewMoney(decrypted.Amount, domain.Currency(decrypted.Currency))
		converted, cErr := h.marketplace.ConvertHistorical(c.Request.Context(), money, displayCurrency, preferredQuote, expectedDate)
		if cErr != nil {
			handleServiceError(c, cErr)
			return
		}
		expectedTotal = expectedTotal.Add(converted.Amount)
	}

	actualTotal := decimal.Zero
	for _, e := range actualExpenses {
		decrypted, dErr := h.encryptor.DecryptMoney(e.EncryptedAmount())
		if dErr != nil {
			handleServiceError(c, dErr)
			return
		}
		money := domain.NewMoney(decrypted.Amount, domain.Currency(decrypted.Currency))
		converted, cErr := h.marketplace.ConvertHistorical(c.Request.Context(), money, displayCurrency, preferredQuote, e.ExpenseDate())
		if cErr != nil {
			handleServiceError(c, cErr)
			return
		}
		actualTotal = actualTotal.Add(converted.Amount)
	}

	diff := expectedTotal.Sub(actualTotal)
	response := BudgetSummaryResponse{
		BudgetID: budgetID,
		ExpectedTotal: MoneyResponse{Amount: expectedTotal.String(), Currency: string(displayCurrency), Converted: true},
		ActualTotal:   MoneyResponse{Amount: actualTotal.String(), Currency: string(displayCurrency), Converted: true},
		Difference:    MoneyResponse{Amount: diff.String(), Currency: string(displayCurrency), Converted: true},
	}

	c.JSON(http.StatusOK, response)
}
