package vehicle

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

func (h *Handler) CreateVehicle(c *gin.Context) {
	var req CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	_, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
		return
	}

	resp, err := h.service.CreateVehicle(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, "vehicle created", resp)
}
