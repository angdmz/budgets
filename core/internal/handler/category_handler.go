package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
)

type CategoryHandler struct {
	pool *pgxpool.Pool
}

func NewCategoryHandler(pool *pgxpool.Pool) *CategoryHandler {
	return &CategoryHandler{pool: pool}
}

// CreateCategory godoc
// @Summary Create a new expense category
// @Description Create a new expense category for a budgeting group
// @Tags categories
// @Accept json
// @Produce json
// @Param group_id path string true "Group ID (UUID)"
// @Param request body CreateCategoryRequest true "Category creation request"
// @Success 201 {object} CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups/{group_id}/categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_group_id", Message: "Invalid UUID format"})
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var response CategoryResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupAccess(ctx, p, groupID); err != nil {
			return err
		}

		persistibleCategory, err := domain.NewPersistibleCategory(req.Name, req.Description, req.Color, req.Icon, groupID)
		if err != nil {
			return err
		}

		persistedCategory, err := persistibleCategory.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		response = CategoryResponse{
			ID:          persistedCategory.ExternalID(),
			Name:        persistedCategory.Name(),
			Description: persistedCategory.Description(),
			Color:       persistedCategory.Color(),
			Icon:        persistedCategory.Icon(),
			CreatedAt:   persistedCategory.CreatedAt(),
			UpdatedAt:   persistedCategory.UpdatedAt(),
		}
		return nil
	})

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

	c.JSON(http.StatusCreated, response)
}

// GetCategories godoc
// @Summary Get all categories for a group
// @Description Get all expense categories for a budgeting group
// @Tags categories
// @Produce json
// @Param group_id path string true "Group ID (UUID)"
// @Success 200 {array} CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups/{group_id}/categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_group_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var response []CategoryResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupAccess(ctx, p, groupID); err != nil {
			return err
		}

		categories, err := domain.PersistedCategoriesForGroup(ctx, groupID, p)
		if err != nil {
			return err
		}

		response = make([]CategoryResponse, len(categories))
		for i, cat := range categories {
			response[i] = CategoryResponse{
				ID:          cat.ExternalID(),
				Name:        cat.Name(),
				Description: cat.Description(),
				Color:       cat.Color(),
				Icon:        cat.Icon(),
				CreatedAt:   cat.CreatedAt(),
				UpdatedAt:   cat.UpdatedAt(),
			}
		}
		return nil
	})

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

	c.JSON(http.StatusOK, response)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update an expense category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID (UUID)"
// @Param request body UpdateCategoryRequest true "Category update request"
// @Success 200 {object} CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var response CategoryResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeCategoryAccess(ctx, p, id); err != nil {
			return err
		}

		category, err := domain.PersistedCategoryFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		category.UpdateName(req.Name)
		category.UpdateDescription(req.Description)
		category.UpdateColor(req.Color)
		category.UpdateIcon(req.Icon)

		if err := category.UpdateIn(ctx, p); err != nil {
			return err
		}

		response = CategoryResponse{
			ID:          category.ExternalID(),
			Name:        category.Name(),
			Description: category.Description(),
			Color:       category.Color(),
			Icon:        category.Icon(),
			CreatedAt:   category.CreatedAt(),
			UpdatedAt:   category.UpdatedAt(),
		}
		return nil
	})

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

	c.JSON(http.StatusOK, response)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Soft delete an expense category by ID
// @Tags categories
// @Produce json
// @Param id path string true "Category ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeCategoryAccess(ctx, p, id); err != nil {
			return err
		}

		category, err := domain.PersistedCategoryFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		return category.DeleteFrom(ctx, p)
	})

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
