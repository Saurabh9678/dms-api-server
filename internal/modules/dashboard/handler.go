package dashboard

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDashboard(c *gin.Context) {
	duration := c.DefaultQuery("duration", "lifetime")

	var showroomID *uint64
	if raw := c.Query("showroom_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
			return
		}
		showroomID = &id
	}

	resp, err := h.service.GetDashboard(c.Request.Context(), GetDashboardRequest{
		Duration:   duration,
		ShowroomID: showroomID,
	})
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, "dashboard data fetched", resp)
}
