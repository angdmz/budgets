package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/service"
)

type BudgetHandler struct {
	budgetService service.BudgetService
}

func NewBudgetHandler(budgetService service.BudgetService) *BudgetHandler {
	return &BudgetHandler{budgetService: budgetService}
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
	groupIDStr := c.Param("group_id")
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

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	budget, err := h.budgetService.Create(c.Request.Context(), groupID, req.Name, req.Description, startDate, endDate, user.ExternalProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "group_not_found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusCreated, BudgetResponse{
		ID:          budget.ExternalID,
		Name:        budget.Name,
		Description: budget.Description,
		StartDate:   budget.StartDate.Format(dateFormat),
		EndDate:     budget.EndDate.Format(dateFormat),
		CreatedAt:   budget.CreatedAt,
		UpdatedAt:   budget.UpdatedAt,
	})
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
	groupIDStr := c.Param("group_id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_group_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	budgets, err := h.budgetService.GetByGroupID(c.Request.Context(), groupID, user.ExternalProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "group_not_found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	response := make([]BudgetResponse, len(budgets))
	for i, b := range budgets {
		response[i] = BudgetResponse{
			ID:          b.ExternalID,
			Name:        b.Name,
			Description: b.Description,
			StartDate:   b.StartDate.Format(dateFormat),
			EndDate:     b.EndDate.Format(dateFormat),
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
		}
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	budget, err := h.budgetService.GetByID(c.Request.Context(), id, user.ExternalProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, BudgetResponse{
		ID:          budget.ExternalID,
		Name:        budget.Name,
		Description: budget.Description,
		StartDate:   budget.StartDate.Format(dateFormat),
		EndDate:     budget.EndDate.Format(dateFormat),
		CreatedAt:   budget.CreatedAt,
		UpdatedAt:   budget.UpdatedAt,
	})
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
	idStr := c.Param("id")
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

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	budget, err := h.budgetService.Update(c.Request.Context(), id, req.Name, req.Description, startDate, endDate, user.ExternalProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, BudgetResponse{
		ID:          budget.ExternalID,
		Name:        budget.Name,
		Description: budget.Description,
		StartDate:   budget.StartDate.Format(dateFormat),
		EndDate:     budget.EndDate.Format(dateFormat),
		CreatedAt:   budget.CreatedAt,
		UpdatedAt:   budget.UpdatedAt,
	})
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
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	err = h.budgetService.Delete(c.Request.Context(), id, user.ExternalProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.Status(http.StatusNoContent)
}
