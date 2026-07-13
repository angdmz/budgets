package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
)

type BudgetHandler struct {
	pool *pgxpool.Pool
}

func NewBudgetHandler(pool *pgxpool.Pool) *BudgetHandler {
	return &BudgetHandler{pool: pool}
}

const dateFormat = "2006-01-02"

// CreateBudget godoc
// @Summary Create a new budget
// @Description Create a new budget for a budgeting group
// @Tags budgets
// @Accept json
// @Produce json
// @Param group_id path string true "Group ID (UUID)"
// @Param request body CreateBudgetRequest true "Budget creation request"
// @Success 201 {object} BudgetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups/{group_id}/budgets [post]
func (h *BudgetHandler) CreateBudget(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_group_id", Message: "Invalid UUID format"})
		return
	}

	var req CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	startDate, err := time.Parse(dateFormat, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_start_date", Message: "Date must be in YYYY-MM-DD format"})
		return
	}

	endDate, err := time.Parse(dateFormat, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_end_date", Message: "Date must be in YYYY-MM-DD format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response BudgetResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupAccess(ctx, p, groupID); err != nil {
			return err
		}

		persistibleBudget, err := domain.NewPersistibleBudget(req.Name, req.Description, startDate, endDate, groupID)
		if err != nil {
			return err
		}

		persistedBudget, err := persistibleBudget.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		response = BudgetResponse{
			ID:          persistedBudget.ExternalID(),
			Name:        persistedBudget.Name(),
			Description: persistedBudget.Description(),
			StartDate:   persistedBudget.StartDate().Format(dateFormat),
			EndDate:     persistedBudget.EndDate().Format(dateFormat),
			CreatedAt:   persistedBudget.CreatedAt(),
			UpdatedAt:   persistedBudget.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "group_not_found", Message: "Group not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to perform this action"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetBudgets godoc
// @Summary Get all budgets for a group
// @Description Get all budgets for a budgeting group
// @Tags budgets
// @Produce json
// @Param group_id path string true "Group ID (UUID)"
// @Success 200 {array} BudgetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups/{group_id}/budgets [get]
func (h *BudgetHandler) GetBudgets(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_group_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response []BudgetResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupAccess(ctx, p, groupID); err != nil {
			return err
		}

		budgets, err := domain.PersistedBudgetsForGroup(ctx, groupID, p)
		if err != nil {
			return err
		}

		response = make([]BudgetResponse, len(budgets))
		for i, b := range budgets {
			response[i] = BudgetResponse{
				ID:          b.ExternalID(),
				Name:        b.Name(),
				Description: b.Description(),
				StartDate:   b.StartDate().Format(dateFormat),
				EndDate:     b.EndDate().Format(dateFormat),
				CreatedAt:   b.CreatedAt(),
				UpdatedAt:   b.UpdatedAt(),
			}
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "group_not_found", Message: "Group not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to perform this action"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBudget godoc
// @Summary Get a specific budget
// @Description Get a budget by ID
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID (UUID)"
// @Success 200 {object} BudgetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{id} [get]
func (h *BudgetHandler) GetBudget(c *gin.Context) {
	idStr := c.Param("budget_id")
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

	var response BudgetResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, id); err != nil {
			return err
		}

		budget, err := domain.PersistedBudgetFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		response = BudgetResponse{
			ID:          budget.ExternalID(),
			Name:        budget.Name(),
			Description: budget.Description(),
			StartDate:   budget.StartDate().Format(dateFormat),
			EndDate:     budget.EndDate().Format(dateFormat),
			CreatedAt:   budget.CreatedAt(),
			UpdatedAt:   budget.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Budget not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to perform this action"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateBudget godoc
// @Summary Update a budget
// @Description Update a budget by ID
// @Tags budgets
// @Accept json
// @Produce json
// @Param id path string true "Budget ID (UUID)"
// @Param request body UpdateBudgetRequest true "Budget update request"
// @Success 200 {object} BudgetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{id} [put]
func (h *BudgetHandler) UpdateBudget(c *gin.Context) {
	idStr := c.Param("budget_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	startDate, err := time.Parse(dateFormat, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_start_date", Message: "Date must be in YYYY-MM-DD format"})
		return
	}

	endDate, err := time.Parse(dateFormat, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_end_date", Message: "Date must be in YYYY-MM-DD format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response BudgetResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeBudgetAccess(ctx, p, id); err != nil {
			return err
		}

		budget, err := domain.PersistedBudgetFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		budget.UpdateName(req.Name)
		budget.UpdateDescription(req.Description)
		if err := budget.UpdateDates(startDate, endDate); err != nil {
			return err
		}

		if err := budget.UpdateIn(ctx, p); err != nil {
			return err
		}

		response = BudgetResponse{
			ID:          budget.ExternalID(),
			Name:        budget.Name(),
			Description: budget.Description(),
			StartDate:   budget.StartDate().Format(dateFormat),
			EndDate:     budget.EndDate().Format(dateFormat),
			CreatedAt:   budget.CreatedAt(),
			UpdatedAt:   budget.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Budget not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to perform this action"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteBudget godoc
// @Summary Delete a budget
// @Description Soft delete a budget by ID
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /budgets/{id} [delete]
func (h *BudgetHandler) DeleteBudget(c *gin.Context) {
	idStr := c.Param("budget_id")
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
		if err := guard.AuthorizeBudgetAccess(ctx, p, id); err != nil {
			return err
		}

		budget, err := domain.PersistedBudgetFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		return budget.DeleteFrom(ctx, p)
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Budget not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to perform this action"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.Status(http.StatusNoContent)
}
