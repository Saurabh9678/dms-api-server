package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

type Handler struct {
	service Service
}

type authHeaders struct {
	Platform string `header:"X-Platform" binding:"required,oneof=web ios_mobile android_mobile desktop"`
	DeviceID string `header:"X-Device-Id"`
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}
	headers, ok := bindAuthHeaders(c)
	if !ok {
		return
	}
	req.Platform = headers.Platform
	req.DeviceID = headers.DeviceID
	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, "OTP sent successfully", resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}
	headers, ok := bindAuthHeaders(c)
	if !ok {
		return
	}
	req.Platform = headers.Platform
	req.DeviceID = headers.DeviceID
	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, "OTP sent successfully", resp)
}

func (h *Handler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}
	headers, ok := bindAuthHeaders(c)
	if !ok {
		return
	}
	req.Platform = headers.Platform
	req.DeviceID = headers.DeviceID
	resp, err := h.service.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, "OTP verified successfully", resp)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}
	resp, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, "Token refreshed successfully", resp)
}

func (h *Handler) Logout(c *gin.Context) {
	accessToken, ok := extractBearerToken(c)
	if !ok {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}
	headers, ok := bindAuthHeaders(c)
	if !ok {
		return
	}
	req := LogoutRequest{AccessToken: accessToken, Platform: headers.Platform}
	if err := h.service.Logout(c.Request.Context(), req); err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, "Logged out successfully", nil)
}

func bindAuthHeaders(c *gin.Context) (*authHeaders, bool) {
	var headers authHeaders
	if err := c.ShouldBindHeader(&headers); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return nil, false
	}
	return &headers, true
}

func extractBearerToken(c *gin.Context) (string, bool) {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return "", false
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if token == "" {
		return "", false
	}

	return token, true
}
