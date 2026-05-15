package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/internal/application/auth"
	"infiour.local/dms-api-server/internal/domain/user"
)

type AuthHandler struct {
	authService AuthService
}

type AuthService interface {
	Register(ctx context.Context, req auth.RegisterRequest) (*auth.TriggerOTPResponse, error)
	Login(ctx context.Context, req auth.LoginRequest) (*auth.TriggerOTPResponse, error)
	VerifyOTP(ctx context.Context, req auth.VerifyOTPRequest) (*auth.TokenResponse, error)
	RefreshToken(ctx context.Context, req auth.RefreshTokenRequest) (*auth.TokenResponse, error)
	Logout(ctx context.Context, req auth.LogoutRequest) error
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req auth.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.authService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req auth.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := h.authService.Logout(c.Request.Context(), req); err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, user.ErrInvalidOTP),
		errors.Is(err, user.ErrOTPExpired),
		errors.Is(err, user.ErrOTPAlreadyUsed),
		errors.Is(err, user.ErrOTPAttemptsExceeded),
		errors.Is(err, user.ErrInvalidRefreshToken),
		errors.Is(err, user.ErrSessionRevoked):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
