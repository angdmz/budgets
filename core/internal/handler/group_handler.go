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

type GroupHandler struct {
	pool *pgxpool.Pool
}

func NewGroupHandler(pool *pgxpool.Pool) *GroupHandler {
	return &GroupHandler{pool: pool}
}

// CreateGroup godoc
// @Summary Create a new budgeting group
// @Description Create a new budgeting group and add the current user as a participant
// @Tags groups
// @Accept json
// @Produce json
// @Param request body CreateGroupRequest true "Group creation request"
// @Success 201 {object} GroupResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var response GroupResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		persistibleGroup, err := domain.NewPersistibleGroup(req.Name, req.Description)
		if err != nil {
			return err
		}

		participant := persistibleGroup.AddParticipant(user.DisplayName, "")
		participant.AddUser(user.ID, "owner", true)

		persistedGroup, err := persistibleGroup.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		response = GroupResponse{
			ID:          persistedGroup.ExternalID(),
			Name:        persistedGroup.Name(),
			Description: persistedGroup.Description(),
			CreatedAt:   persistedGroup.CreatedAt(),
			UpdatedAt:   persistedGroup.UpdatedAt(),
		}
		return nil
	})

	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetGroups godoc
// @Summary Get all groups for current user
// @Description Get all budgeting groups the current user is a participant of
// @Tags groups
// @Produce json
// @Success 200 {array} GroupResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups [get]
func (h *GroupHandler) GetGroups(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var response []GroupResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		groups, err := domain.PersistedGroupsForUser(ctx, user.ID, p)
		if err != nil {
			return err
		}

		response = make([]GroupResponse, len(groups))
		for i, g := range groups {
			response[i] = GroupResponse{
				ID:          g.ExternalID(),
				Name:        g.Name(),
				Description: g.Description(),
				CreatedAt:   g.CreatedAt(),
				UpdatedAt:   g.UpdatedAt(),
			}
		}
		return nil
	})

	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetGroup godoc
// @Summary Get a specific group
// @Description Get a budgeting group by ID
// @Tags groups
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Success 200 {object} GroupResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups/{id} [get]
func (h *GroupHandler) GetGroup(c *gin.Context) {
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

	var response GroupResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupAccess(ctx, p, id); err != nil {
			return err
		}

		group, err := domain.PersistedGroupFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		response = GroupResponse{
			ID:          group.ExternalID(),
			Name:        group.Name(),
			Description: group.Description(),
			CreatedAt:   group.CreatedAt(),
			UpdatedAt:   group.UpdatedAt(),
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

// UpdateGroup godoc
// @Summary Update a group
// @Description Update a budgeting group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Param request body UpdateGroupRequest true "Group update request"
// @Success 200 {object} GroupResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups/{id} [put]
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	var response GroupResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupAccess(ctx, p, id); err != nil {
			return err
		}

		group, err := domain.PersistedGroupFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		group.UpdateName(req.Name)
		group.UpdateDescription(req.Description)

		if err := group.UpdateIn(ctx, p); err != nil {
			return err
		}

		response = GroupResponse{
			ID:          group.ExternalID(),
			Name:        group.Name(),
			Description: group.Description(),
			CreatedAt:   group.CreatedAt(),
			UpdatedAt:   group.UpdatedAt(),
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

// DeleteGroup godoc
// @Summary Delete a group
// @Description Soft delete a budgeting group by ID
// @Tags groups
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
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
		if err := guard.AuthorizeGroupAccess(ctx, p, id); err != nil {
			return err
		}

		group, err := domain.PersistedGroupFromPersistence(ctx, id, p)
		if err != nil {
			return err
		}

		return group.DeleteFrom(ctx, p)
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
