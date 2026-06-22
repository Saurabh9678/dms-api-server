package showroom

import (
	"mime/multipart"
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

func (h *Handler) CreateShowroom(c *gin.Context) {
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

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	req := &CreateShowroomRequest{
		Name:        c.Request.FormValue("name"),
		Geolocation: c.Request.FormValue("geolocation"),
	}

	var logo, banner *multipart.FileHeader
	form := c.Request.MultipartForm
	if logoFiles := form.File["showroom_logo"]; len(logoFiles) > 0 {
		logo = logoFiles[0]
	}
	if bannerFiles := form.File["showroom_banner"]; len(bannerFiles) > 0 {
		banner = bannerFiles[0]
	}

	resp, err := h.service.CreateShowroom(c.Request.Context(), userID, req, logo, banner)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, "showroom created", resp)
}
