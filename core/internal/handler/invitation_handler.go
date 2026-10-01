package handler

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/representation"
)

type InvitationHandler struct {
	pool *pgxpool.Pool
}

func NewInvitationHandler(pool *pgxpool.Pool) *InvitationHandler {
	return &InvitationHandler{pool: pool}
}

func (h *InvitationHandler) CreateInvitation(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Role = "member"
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response domain.Rendered[representation.Invitation]
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		if _, err := domain.PersistedGroupFromPersistence(ctx, groupID, p); err != nil {
			return err
		}

		guard := middleware.NewSecurityGuard(user)
		if err := guard.AuthorizeGroupOwnership(ctx, p, groupID); err != nil {
			return err
		}

		invitation, err := domain.NewPersistibleInvitation(groupID, user, domain.ParticipantRole(req.Role), rand.Reader, time.Now())
		if err != nil {
			return err
		}

		persisted, err := invitation.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		response = persisted.Render()
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "group_not_found", Message: "Group not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to manage invitations for this group"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *InvitationHandler) ListInvitations(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid UUID format"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response []domain.Rendered[representation.Invitation]
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		if _, err := domain.PersistedGroupFromPersistence(ctx, groupID, p); err != nil {
			return err
		}

		guard := middleware.NewSecurityGuard(user)
		if err := guard.AuthorizeGroupOwnership(ctx, p, groupID); err != nil {
			return err
		}

		invitations, err := domain.PersistedInvitationsForGroup(ctx, groupID, p)
		if err != nil {
			return err
		}

		response = make([]domain.Rendered[representation.Invitation], len(invitations))
		for i := range invitations {
			response[i] = invitations[i].Render()
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "group_not_found", Message: "Group not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to manage invitations for this group"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InvitationHandler) RevokeInvitation(c *gin.Context) {
	idStr := c.Param("id")
	invitationID, err := uuid.Parse(idStr)
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
		// Authorize before loading the resource: nothing is fetched until the
		// caller's ownership is proven.
		guard := middleware.NewSecurityGuard(user)
		if err := guard.AuthorizeInvitationOwnership(ctx, p, invitationID); err != nil {
			return err
		}

		invitation, err := domain.PersistedInvitationByExternalID(ctx, invitationID, p)
		if err != nil {
			return err
		}

		return invitation.Revoke(ctx, p)
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Invitation not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: "You do not have permission to manage invitations for this group"})
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "conflict", Message: "Cannot revoke invitation in current state"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *InvitationHandler) GetInvitationByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_token", Message: "Invitation token is required"})
		return
	}

	var response domain.Rendered[representation.InvitationDetail]
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		invitation, err := domain.PersistedInvitationByToken(ctx, token, p)
		if err != nil {
			return err
		}

		if err := invitation.EnsureUsable(time.Now()); err != nil {
			return err
		}

		response = invitation.RenderDetail()
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Invitation not found"})
			return
		}
		if errors.Is(err, domain.ErrGone) {
			c.JSON(http.StatusGone, ErrorResponse{Error: "invitation_expired_or_revoked", Message: "This invitation has expired or been revoked"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_token", Message: "Invitation token is required"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		invitation, err := domain.PersistedInvitationByToken(ctx, token, p)
		if err != nil {
			return err
		}

		return invitation.Accept(ctx, user, time.Now(), p)
	})

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Invitation not found"})
			return
		}
		if errors.Is(err, domain.ErrConflict) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "conflict", Message: "Invitation already used or user already a member"})
			return
		}
		if errors.Is(err, domain.ErrGone) {
			c.JSON(http.StatusGone, ErrorResponse{Error: "invitation_expired", Message: "This invitation has expired"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted successfully"})
}
