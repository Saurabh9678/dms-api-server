package showroom

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/middleware"
	"infiour.local/dms-api-server/pkg/response"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateShowroom(c *gin.Context) {
	userID, ok := h.extractUserID(c)
	if !ok {
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

func (h *Handler) AddMember(c *gin.Context) {
	showroomID, ok := h.parseShowroomID(c)
	if !ok {
		return
	}

	callerRoles, ok := h.extractShowroomRoles(c)
	if !ok {
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	resp, err := h.service.AddMember(c.Request.Context(), callerRoles, showroomID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, "member added", resp)
}

func (h *Handler) ListMembers(c *gin.Context) {
	showroomID, ok := h.parseShowroomID(c)
	if !ok {
		return
	}

	callerRoles, ok := h.extractShowroomRoles(c)
	if !ok {
		return
	}

	page := parseIntParam(c.Query("page"), defaultPage, 1, 0)
	limit := parseIntParam(c.Query("limit"), defaultLimit, 1, maxLimit)

	resp, err := h.service.ListMembers(c.Request.Context(), callerRoles, showroomID, page, limit)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "members fetched", resp)
}

func (h *Handler) RemoveMember(c *gin.Context) {
	callerUserID, ok := h.extractUserID(c)
	if !ok {
		return
	}

	showroomID, ok := h.parseShowroomID(c)
	if !ok {
		return
	}

	targetUserID, ok := h.parseTargetUserID(c)
	if !ok {
		return
	}

	callerRoles, ok := h.extractShowroomRoles(c)
	if !ok {
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), callerUserID, callerRoles, showroomID, targetUserID); err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "member removed", nil)
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	callerUserID, ok := h.extractUserID(c)
	if !ok {
		return
	}

	showroomID, ok := h.parseShowroomID(c)
	if !ok {
		return
	}

	targetUserID, ok := h.parseTargetUserID(c)
	if !ok {
		return
	}

	callerRoles, ok := h.extractShowroomRoles(c)
	if !ok {
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	resp, err := h.service.UpdateMemberRole(c.Request.Context(), callerUserID, callerRoles, showroomID, targetUserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "member role updated", resp)
}

func (h *Handler) extractUserID(c *gin.Context) (uint64, bool) {
	val, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
		return 0, false
	}
	userID, ok := val.(uint64)
	if !ok {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
		return 0, false
	}
	return userID, true
}

func (h *Handler) extractShowroomRoles(c *gin.Context) (map[uint64]string, bool) {
	val, exists := c.Get(middleware.ContextKeyShowroomRoles)
	if !exists {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
		return nil, false
	}
	roles, ok := val.(map[uint64]string)
	if !ok {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
		return nil, false
	}
	return roles, true
}

func (h *Handler) parseShowroomID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return 0, false
	}
	return id, true
}

func (h *Handler) parseTargetUserID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return 0, false
	}
	return id, true
}

// parseIntParam parses a query string integer with a default, minimum, and optional maximum (0 = no max).
func parseIntParam(raw string, def, min, max int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < min {
		return def
	}
	if max > 0 && v > max {
		return max
	}
	return v
}
