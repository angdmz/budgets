package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/budgets/core/internal/domain"
	"github.com/budgets/core/internal/middleware"
)

type AuthHandler struct {
	oauthConfig    *oauth2.Config
	authMiddleware *middleware.AuthMiddleware
}

func NewAuthHandler(clientID, clientSecret, redirectURL string, authMiddleware *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		authMiddleware: authMiddleware,
	}
}

// GoogleLogin godoc
// @Summary Initiate Google OAuth login
// @Description Redirects to Google OAuth consent screen
// @Tags auth
// @Success 302
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := "random-state-string"
	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback godoc
// @Summary Handle Google OAuth callback
// @Description Exchanges auth code for token and returns JWT
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} AuthCallbackResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "missing_code"})
		return
	}

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "token_exchange_failed", err)
		return
	}

	userInfo, err := h.getUserInfo(token.AccessToken)
	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "user_info_failed", err)
		return
	}

	user := &domain.User{
		ExternalProviderID: userInfo.ID,
		Email:              userInfo.Email,
		DisplayName:        userInfo.Name,
		AuthProvider:       domain.AuthProviderGoogle,
	}

	jwtToken, err := h.authMiddleware.GenerateToken(user)
	if err != nil {
		SafeErrorResponse(c, http.StatusInternalServerError, "token_generation_failed", err)
		return
	}

	c.JSON(http.StatusOK, AuthCallbackResponse{Token: jwtToken})
}

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (h *AuthHandler) getUserInfo(accessToken string) (*googleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userInfo googleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}

// GetCurrentUser godoc
// @Summary Get current user info
// @Description Returns the current authenticated user's information
// @Tags auth
// @Produce json
// @Success 200 {object} domain.User
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user := middleware.GetDBUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, user)
}
