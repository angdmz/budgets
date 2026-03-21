package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/service"
)

type GroupHandler struct {
	groupService service.GroupService
}

func NewGroupHandler(groupService service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
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

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	group, err := h.groupService.Create(c.Request.Context(), req.Name, req.Description, user.ExternalProviderID, user.Email, user.DisplayName)
	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusCreated, GroupResponse{
		ID:          group.ExternalID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	})
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
	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	groups, err := h.groupService.GetByUserID(c.Request.Context(), user.ExternalProviderID)
	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	response := make([]GroupResponse, len(groups))
	for i, g := range groups {
		response[i] = GroupResponse{
			ID:          g.ExternalID,
			Name:        g.Name,
			Description: g.Description,
			CreatedAt:   g.CreatedAt,
			UpdatedAt:   g.UpdatedAt,
		}
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

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	group, err := h.groupService.GetByID(c.Request.Context(), id, user.ExternalProviderID)
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

	c.JSON(http.StatusOK, GroupResponse{
		ID:          group.ExternalID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	})
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

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	group, err := h.groupService.Update(c.Request.Context(), id, req.Name, req.Description, user.ExternalProviderID)
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

	c.JSON(http.StatusOK, GroupResponse{
		ID:          group.ExternalID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	})
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

	user := middleware.GetAuth0UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	err = h.groupService.Delete(c.Request.Context(), id, user.ExternalProviderID)
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
