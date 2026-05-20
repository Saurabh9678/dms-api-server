package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/middleware"
	"infiour.local/dms-api-server/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	userIDVal, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
		return
	}

	userID, ok := userIDVal.(uint64)
	if !ok {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
		return
	}

	resp, err := h.service.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "profile updated", resp)
}
