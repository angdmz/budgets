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

	var response InvitationResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		var groupInternalID int64
		err := p.QueryRow(
			ctx,
			[]any{&groupInternalID},
			`SELECT id FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL`,
			groupID,
		)
		if err != nil {
			return domain.ErrNotFound
		}

		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupOwnership(ctx, p, groupID); err != nil {
			return err
		}

		invitation, err := domain.NewPersistibleInvitation(groupInternalID, user.ID, req.Role)
		if err != nil {
			return err
		}

		persisted, err := invitation.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		response = InvitationResponse{
			ID:          persisted.ExternalID(),
			Token:       persisted.Token(),
			GroupName:   persisted.GroupName(),
			InviterName: persisted.InviterName(),
			Status:      persisted.Status(),
			Role:        persisted.Role(),
			ExpiresAt:   persisted.ExpiresAt(),
			AcceptedAt:  persisted.AcceptedAt(),
			CreatedAt:   persisted.CreatedAt(),
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

	var response []InvitationResponse
	err = database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		var groupExists bool
		err := p.QueryRow(
			ctx,
			[]any{&groupExists},
			`SELECT EXISTS(SELECT 1 FROM budgeting_groups WHERE external_id = $1 AND revoked_at IS NULL)`,
			groupID,
		)
		if err != nil || !groupExists {
			return domain.ErrNotFound
		}

		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupOwnership(ctx, p, groupID); err != nil {
			return err
		}

		invitations, err := domain.PersistedInvitationsForGroup(ctx, groupID, p)
		if err != nil {
			return err
		}

		response = make([]InvitationResponse, len(invitations))
		for i, inv := range invitations {
			response[i] = InvitationResponse{
				ID:          inv.ExternalID(),
				Token:       inv.Token(),
				GroupName:   inv.GroupName(),
				InviterName: inv.InviterName(),
				Status:      inv.Status(),
				Role:        inv.Role(),
				ExpiresAt:   inv.ExpiresAt(),
				AcceptedAt:  inv.AcceptedAt(),
				CreatedAt:   inv.CreatedAt(),
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
		invitation, err := domain.PersistedInvitationByExternalID(ctx, invitationID, p)
		if err != nil {
			return err
		}

		var groupExternalID uuid.UUID
		err = p.QueryRow(
			ctx,
			[]any{&groupExternalID},
			`SELECT external_id FROM budgeting_groups WHERE id = $1`,
			invitation.GroupID(),
		)
		if err != nil {
			return err
		}

		guard := domain.NewSecurityGuard(user.ID)
		if err := guard.AuthorizeGroupOwnership(ctx, p, groupExternalID); err != nil {
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

	var response InvitationDetailResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		invitation, err := domain.PersistedInvitationByToken(ctx, token, p)
		if err != nil {
			return err
		}

		if invitation.IsExpired() || invitation.Status() == domain.InvitationStatusRevoked {
			return domain.ErrGone
		}

		response = InvitationDetailResponse{
			GroupName:   invitation.GroupName(),
			InviterName: invitation.InviterName(),
			Status:      invitation.Status(),
			Role:        invitation.Role(),
			ExpiresAt:   invitation.ExpiresAt(),
		}
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

		return invitation.Accept(ctx, user.ID, p)
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
