package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"finance-tracker/pkg/apperror"
	"finance-tracker/pkg/auth"
	"finance-tracker/pkg/middleware"
	"finance-tracker/pkg/models"
	"finance-tracker/pkg/service"

	"github.com/gin-gonic/gin"
)

const (
	refreshTokenCookieName = "refresh_token"
	csrfTokenCookieName    = "csrf_token"
)

type authService interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.AuthTokens, *apperror.Error)
	Login(ctx context.Context, req models.LoginRequest) (*models.AuthTokens, *apperror.Error)
	Refresh(ctx context.Context, rawRefreshToken string) (*models.AuthTokens, *apperror.Error)
	Logout(ctx context.Context, userID int64, rawRefreshToken, rawAccessToken string) *apperror.Error
}

type AuthHandler struct {
	authService authService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	if authService == nil {
		return &AuthHandler{}
	}
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register
// @Description Register a new user and return access/refresh tokens.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register payload"
// @Success 201 {object} AccessTokenResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 409 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.authService.Register(c.Request.Context(), req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	if err := setAuthCookies(c, out.RefreshToken); err != nil {
		writeError(c, apperror.Internal("failed to set auth cookies"))
		return
	}
	c.JSON(http.StatusCreated, safeAuthResponse(out.AccessToken, gin.H{"expires_in": out.ExpiresIn}))
}

// Login godoc
// @Summary Login
// @Description Login with email and password.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Success 200 {object} AccessTokenResponse
// @Failure 400 {object} ErrorEnvelope
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, apperror.Validation(err.Error()))
		return
	}
	out, appErr := h.authService.Login(c.Request.Context(), req)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	if err := setAuthCookies(c, out.RefreshToken); err != nil {
		writeError(c, apperror.Internal("failed to set auth cookies"))
		return
	}
	c.JSON(http.StatusOK, safeAuthResponse(out.AccessToken, gin.H{"expires_in": out.ExpiresIn}))
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Rotate refresh token and return new tokens.
// @Tags auth
// @Produce json
// @Param Cookie header string true "Cookie header containing refresh_token=<token>"
// @Success 200 {object} AccessTokenResponse
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := readRefreshTokenFromCookie(c)
	if err != nil {
		writeError(c, apperror.Unauthorized("missing refresh token cookie"))
		return
	}
	out, appErr := h.authService.Refresh(c.Request.Context(), refreshToken)
	if appErr != nil {
		writeError(c, appErr)
		return
	}
	if err = setAuthCookies(c, out.RefreshToken); err != nil {
		writeError(c, apperror.Internal("failed to set auth cookies"))
		return
	}
	c.JSON(http.StatusOK, safeAuthResponse(out.AccessToken, gin.H{"expires_in": out.ExpiresIn}))
}

// Logout godoc
// @Summary Logout
// @Description Revoke current refresh token.
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Param Cookie header string true "Cookie header containing refresh_token=<token>"
// @Success 204 {string} string "No Content"
// @Failure 401 {object} ErrorEnvelope
// @Failure 500 {object} ErrorEnvelope
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, apperror.Unauthorized("invalid token context"))
		return
	}
	accessToken, authErr := middleware.AccessTokenFromHeader(c.GetHeader("Authorization"))
	if authErr != nil {
		writeError(c, authErr)
		return
	}
	refreshToken, err := readRefreshTokenFromCookie(c)
	if err != nil {
		writeError(c, apperror.Unauthorized("missing refresh token cookie"))
		return
	}
	if appErr := h.authService.Logout(c.Request.Context(), userID, refreshToken, accessToken); appErr != nil {
		writeError(c, appErr)
		return
	}
	clearAuthCookies(c)
	c.Status(http.StatusNoContent)
}

func safeAuthResponse(accessToken string, payload gin.H) gin.H {
	out := gin.H{}
	for k, v := range payload {
		out[k] = v
	}
	out["access_token"] = accessToken
	for _, k := range []string{"refresh_token", "refresh_session_id", "session_id", "jti", "token_id"} {
		delete(out, k)
	}
	return out
}

func setAuthCookies(c *gin.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return errors.New("empty refresh token")
	}
	sameSite := parseCookieSameSite(getenv("COOKIE_SAMESITE", "strict"))
	secure := cookieSecure(c)
	maxAge := int(auth.RefreshTokenTTL / time.Second)

	c.SetSameSite(sameSite)
	c.SetCookie(refreshTokenCookieName, refreshToken, maxAge, "/", "", secure, true)

	csrfToken, err := auth.GenerateCSRFToken()
	if err != nil {
		return err
	}
	c.SetSameSite(sameSite)
	c.SetCookie(csrfTokenCookieName, csrfToken, maxAge, "/", "", secure, false)
	return nil
}

func clearAuthCookies(c *gin.Context) {
	sameSite := parseCookieSameSite(getenv("COOKIE_SAMESITE", "strict"))
	secure := cookieSecure(c)
	c.SetSameSite(sameSite)
	c.SetCookie(refreshTokenCookieName, "", -1, "/", "", secure, true)
	c.SetSameSite(sameSite)
	c.SetCookie(csrfTokenCookieName, "", -1, "/", "", secure, false)
}

func readRefreshTokenFromCookie(c *gin.Context) (string, error) {
	raw, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", errors.New("empty refresh token cookie")
	}
	return v, nil
}

func parseCookieSameSite(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteStrictMode
	}
}

func cookieSecure(c *gin.Context) bool {
	override := strings.TrimSpace(strings.ToLower(os.Getenv("COOKIE_SECURE")))
	if override != "" {
		parsed, err := strconv.ParseBool(override)
		if err == nil {
			return parsed
		}
	}
	return c.Request != nil && c.Request.TLS != nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
