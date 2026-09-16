package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
)

type PreferenceHandler struct {
	pool *pgxpool.Pool
}

func NewPreferenceHandler(pool *pgxpool.Pool) *PreferenceHandler {
	return &PreferenceHandler{pool: pool}
}

// GetPreferences godoc
// @Summary Get user preferences
// @Description Get the current user's theme, language, and display currency preferences
// @Tags preferences
// @Produce json
// @Success 200 {object} PreferenceResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /preferences [get]
func (h *PreferenceHandler) GetPreferences(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response PreferenceResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		pref, err := domain.PersistedUserPreferenceFromPersistence(ctx, user.ID, p)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			persistible, err := domain.NewPersistibleUserPreference(user.ID, domain.ThemeLight, domain.LanguageEN, domain.CurrencyUSD, domain.QuoteOfficial)
			if err != nil {
				return err
			}
			pref, err = persistible.PersistTo(ctx, p)
			if err != nil {
				return err
			}
		}

		response = PreferenceResponse{
			Theme:              string(pref.Theme()),
			Language:           string(pref.Language()),
			DisplayCurrency:    string(pref.DisplayCurrency()),
			PreferredQuoteType: string(pref.PreferredQuoteType()),
		}
		return nil
	})

	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdatePreferences godoc
// @Summary Update user preferences
// @Description Update the current user's theme, language, and display currency preferences
// @Tags preferences
// @Accept json
// @Produce json
// @Param request body UpdatePreferenceRequest true "Preference update request"
// @Success 200 {object} PreferenceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /preferences [put]
func (h *PreferenceHandler) UpdatePreferences(c *gin.Context) {
	var req UpdatePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	quoteType := domain.QuoteType(req.PreferredQuoteType)
	if quoteType == "" {
		quoteType = domain.QuoteOfficial
	}

	var response PreferenceResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		persistible, err := domain.NewPersistibleUserPreference(user.ID, domain.Theme(req.Theme), domain.Language(req.Language), domain.Currency(req.DisplayCurrency), quoteType)
		if err != nil {
			return err
		}

		pref, err := persistible.PersistTo(ctx, p)
		if err != nil {
			return err
		}

		response = PreferenceResponse{
			Theme:              string(pref.Theme()),
			Language:           string(pref.Language()),
			DisplayCurrency:    string(pref.DisplayCurrency()),
			PreferredQuoteType: string(pref.PreferredQuoteType()),
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_preference", Message: "Invalid theme, language, or currency value"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// PatchPreferences godoc
// @Summary Partially update user preferences
// @Description Update one or more of the current user's theme, language, and display currency preferences
// @Tags preferences
// @Accept json
// @Produce json
// @Param request body PatchPreferenceRequest true "Preference patch request"
// @Success 200 {object} PreferenceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /preferences [patch]
func (h *PreferenceHandler) PatchPreferences(c *gin.Context) {
	var req PatchPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeValidationError(c, err)
		return
	}

	if req.Theme == nil && req.Language == nil && req.DisplayCurrency == nil && req.PreferredQuoteType == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "empty_request", Message: "At least one field must be provided"})
		return
	}

	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var response PreferenceResponse
	err := database.WithPersister(c.Request.Context(), h.pool, func(ctx context.Context, p *database.PgxPersister) error {
		pref, err := domain.PersistedUserPreferenceFromPersistence(ctx, user.ID, p)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				return err
			}
			persistible, err := domain.NewPersistibleUserPreference(user.ID, domain.ThemeLight, domain.LanguageEN, domain.CurrencyUSD, domain.QuoteOfficial)
			if err != nil {
				return err
			}
			pref, err = persistible.PersistTo(ctx, p)
			if err != nil {
				return err
			}
		}

		if req.Theme != nil {
			if err := pref.UpdateTheme(domain.Theme(*req.Theme)); err != nil {
				return err
			}
		}
		if req.Language != nil {
			if err := pref.UpdateLanguage(domain.Language(*req.Language)); err != nil {
				return err
			}
		}
		if req.DisplayCurrency != nil {
			if err := pref.UpdateDisplayCurrency(domain.Currency(*req.DisplayCurrency)); err != nil {
				return err
			}
		}
		if req.PreferredQuoteType != nil {
			if err := pref.UpdatePreferredQuoteType(domain.QuoteType(*req.PreferredQuoteType)); err != nil {
				return err
			}
		}

		if err := pref.UpdateIn(ctx, p); err != nil {
			return err
		}

		response = PreferenceResponse{
			Theme:              string(pref.Theme()),
			Language:           string(pref.Language()),
			DisplayCurrency:    string(pref.DisplayCurrency()),
			PreferredQuoteType: string(pref.PreferredQuoteType()),
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_preference", Message: "Invalid theme, language, or currency value"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
