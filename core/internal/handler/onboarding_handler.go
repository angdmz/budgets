package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
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

	var response OnboardingResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFromPersistence(ctx, user.ID, p)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			persistible, err := domain.NewPersistibleUserOnboarding(user.ID)
			if err != nil {
				return err
			}
			ob, err = persistible.PersistTo(ctx, p)
			if err != nil {
				return err
			}
		}

		return buildOnboardingResponse(ctx, ob, p, &response)
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

	var response OnboardingResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFromPersistence(ctx, user.ID, p)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			persistible, err := domain.NewPersistibleUserOnboarding(user.ID)
			if err != nil {
				return err
			}
			ob, err = persistible.PersistTo(ctx, p)
			if err != nil {
				return err
			}
		}

		var dataBytes []byte
		if req.Data != nil {
			dataBytes, err = json.Marshal(req.Data)
			if err != nil {
				return err
			}
		}

		persistibleStep, err := domain.NewPersistibleUserOnboardingStep(ob.ID(), step, dataBytes)
		if err != nil {
			return err
		}

		_, err = persistibleStep.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		// Advance to next step or mark onboarding as completed
		if nextStep, hasNext := step.Next(); hasNext {
			if err := ob.AdvanceTo(ctx, nextStep, p); err != nil {
				return err
			}
		} else {
			if err := ob.MarkCompleted(ctx, p); err != nil {
				return err
			}
		}

		return buildOnboardingResponse(ctx, ob, p, &response)
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

	var response OnboardingResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFromPersistence(ctx, user.ID, p)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			persistible, err := domain.NewPersistibleUserOnboarding(user.ID)
			if err != nil {
				return err
			}
			ob, err = persistible.PersistTo(ctx, p)
			if err != nil {
				return err
			}
		}

		persistibleStep, err := domain.NewSkippedUserOnboardingStep(ob.ID(), step)
		if err != nil {
			return err
		}

		_, err = persistibleStep.PersistSkippedTo(ctx, p)
		if err != nil {
			return err
		}

		if nextStep, hasNext := step.Next(); hasNext {
			if err := ob.AdvanceTo(ctx, nextStep, p); err != nil {
				return err
			}
		} else {
			if err := ob.MarkCompleted(ctx, p); err != nil {
				return err
			}
		}

		return buildOnboardingResponse(ctx, ob, p, &response)
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

	var response OnboardingResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFromPersistence(ctx, user.ID, p)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			persistible, err := domain.NewPersistibleUserOnboarding(user.ID)
			if err != nil {
				return err
			}
			ob, err = persistible.PersistTo(ctx, p)
			if err != nil {
				return err
			}
		}

		prevStep, hasPrev := step.Prev()
		if !hasPrev {
			return fmt.Errorf("%w: cannot go back from the first step", domain.ErrValidation)
		}

		if err := ob.AdvanceTo(ctx, prevStep, p); err != nil {
			return err
		}

		return buildOnboardingResponse(ctx, ob, p, &response)
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

func buildOnboardingResponse(ctx context.Context, ob *domain.PersistedUserOnboarding, p *database.PgxPersister, response *OnboardingResponse) error {
	steps, err := ob.Steps(ctx, p)
	if err != nil {
		return err
	}

	stepResponses := make([]OnboardingStepResponse, len(steps))
	for i, s := range steps {
		var data interface{}
		if len(s.Data()) > 0 {
			_ = json.Unmarshal(s.Data(), &data)
		}
		stepResponses[i] = OnboardingStepResponse{
			Step:      string(s.Step()),
			Status:    string(s.Status()),
			Data:      data,
			CreatedAt: s.CreatedAt(),
			UpdatedAt: s.UpdatedAt(),
		}
	}

	*response = OnboardingResponse{
		ID:          ob.ExternalID(),
		Status:      string(ob.Status()),
		CurrentStep: string(ob.CurrentStep()),
		Steps:       stepResponses,
		CreatedAt:   ob.CreatedAt(),
		UpdatedAt:   ob.UpdatedAt(),
	}
	return nil
}

func (h *OnboardingHandler) ResetOnboarding(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response OnboardingResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		ob, err := domain.PersistedUserOnboardingFromPersistence(ctx, user.ID, p)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			persistible, err := domain.NewPersistibleUserOnboarding(user.ID)
			if err != nil {
				return err
			}
			ob, err = persistible.PersistTo(ctx, p)
			if err != nil {
				return err
			}
		}

		if err := ob.Reset(ctx, p); err != nil {
			return err
		}

		return buildOnboardingResponse(ctx, ob, p, &response)
	})

	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
