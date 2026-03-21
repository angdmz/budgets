package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
	"github.com/budgets/core/internal/service"
)

type PreferenceHandler struct {
	preferenceService service.PreferenceService
}

func NewPreferenceHandler(preferenceService service.PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{preferenceService: preferenceService}
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	pref, err := h.preferenceService.Get(c.Request.Context(), user.ID)
	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, PreferenceResponse{
		Theme:           string(pref.Theme),
		Language:        string(pref.Language),
		DisplayCurrency: string(pref.DisplayCurrency),
	})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	pref, err := h.preferenceService.Update(
		c.Request.Context(),
		user.ID,
		domain.Theme(req.Theme),
		domain.Language(req.Language),
		domain.Currency(req.DisplayCurrency),
	)
	if err != nil {
		if err == domain.ErrValidation {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_preference", Message: "Invalid theme, language, or currency value"})
			return
		}
		SafeErrorResponse(c, http.StatusInternalServerError, "internal_error", err)
		return
	}

	c.JSON(http.StatusOK, PreferenceResponse{
		Theme:           string(pref.Theme),
		Language:        string(pref.Language),
		DisplayCurrency: string(pref.DisplayCurrency),
	})
}
