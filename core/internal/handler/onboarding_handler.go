package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/representation"
)

type OnboardingHandler struct {
	pool *pgxpool.Pool
}

func NewOnboardingHandler(pool *pgxpool.Pool) *OnboardingHandler {
	return &OnboardingHandler{pool: pool}
}

func (h *OnboardingHandler) GetOnboarding(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response domain.Rendered[representation.Onboarding]
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFor(ctx, user, p)
		if err != nil {
			return err
		}

		detail, err := ob.Detail(ctx, p)
		if err != nil {
			return err
		}
		response = detail.Render()
		return nil
	})

	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *OnboardingHandler) CompleteStep(c *gin.Context) {
	stepStr := c.Param("step")
	step := domain.OnboardingStep(stepStr)
	if !step.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_step", Message: "Unknown onboarding step"})
		return
	}

	var req CompleteStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if err.Error() == "EOF" {
			req.Data = nil
		} else {
			SafeValidationError(c, err)
			return
		}
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response domain.Rendered[representation.Onboarding]
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFor(ctx, user, p)
		if err != nil {
			return err
		}

		var dataBytes []byte
		if req.Data != nil {
			dataBytes, err = json.Marshal(req.Data)
			if err != nil {
				return err
			}
		}

		detail, err := ob.CompleteStep(ctx, step, dataBytes, p)
		if err != nil {
			return err
		}
		response = detail.Render()
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_step_data", Message: err.Error()})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *OnboardingHandler) SkipStep(c *gin.Context) {
	stepStr := c.Param("step")
	step := domain.OnboardingStep(stepStr)
	if !step.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_step", Message: "Unknown onboarding step"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response domain.Rendered[representation.Onboarding]
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFor(ctx, user, p)
		if err != nil {
			return err
		}

		detail, err := ob.SkipStep(ctx, step, p)
		if err != nil {
			return err
		}
		response = detail.Render()
		return nil
	})

	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *OnboardingHandler) GoBack(c *gin.Context) {
	stepStr := c.Param("step")
	step := domain.OnboardingStep(stepStr)
	if !step.IsValid() {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_step", Message: "Unknown onboarding step"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response domain.Rendered[representation.Onboarding]
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFor(ctx, user, p)
		if err != nil {
			return err
		}

		detail, err := ob.GoBackFrom(ctx, step, p)
		if err != nil {
			return err
		}
		response = detail.Render()
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_step_data", Message: err.Error()})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *OnboardingHandler) ResetOnboarding(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response domain.Rendered[representation.Onboarding]
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFor(ctx, user, p)
		if err != nil {
			return err
		}

		if err := ob.Reset(ctx, p); err != nil {
			return err
		}

		detail, err := ob.Detail(ctx, p)
		if err != nil {
			return err
		}
		response = detail.Render()
		return nil
	})

	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
