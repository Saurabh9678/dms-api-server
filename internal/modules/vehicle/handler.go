package vehicle

import (
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) PublicListVehicles(c *gin.Context) {
	var query PublicListVehiclesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	resp, err := h.service.PublicListVehicles(c.Request.Context(), &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "vehicle listing", resp)
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

func (h *Handler) ListVehicles(c *gin.Context) {
	var query ListVehiclesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	resp, err := h.service.ListVehicles(c.Request.Context(), &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "vehicle listing", resp)
}

func (h *Handler) GetVehicle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	details, err := h.service.GetVehicleByID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	showroomRolesVal, exists := c.Get(middleware.ContextKeyShowroomRoles)
	if !exists {
		response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
		return
	}
	showroomRoles, ok := showroomRolesVal.(map[uint64]string)
	if !ok {
		response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
		return
	}

	role, isMember := showroomRoles[details.ShowroomID]
	if !isMember {
		response.Error(c, http.StatusNotFound, apperrors.CodeVehicleNotFound, "vehicle not found")
		return
	}

	if role == "owner" {
		response.OK(c, "vehicle details", buildAdminResponse(details))
	} else {
		response.OK(c, "vehicle details", buildBasicResponse(details))
	}
}

func buildAdminResponse(d *VehicleFullDetails) GetVehicleAdminResponse {
	basic := buildBasicSection(d)

	resp := GetVehicleAdminResponse{
		Basic:     basic,
		Status:    buildStatusSection(d.Statuses),
		Expenses:  buildExpenseItems(d.Expenses),
		Documents: buildDocumentItems(d.Documents),
		Images:    buildImageItems(d.Images),
	}

	if d.Pricing != nil {
		resp.BuyingDetails = &VehicleBuyingSection{
			BuyingPrice: d.Pricing.BuyingPrice,
			BuyingDate:  d.Pricing.BuyingDate.Format(time.RFC3339),
			Currency:    string(d.Pricing.Currency),
			Remarks:     d.Pricing.Remarks,
		}
		resp.Pricing = &VehiclePricingSection{
			PriceTag: d.Pricing.PriceTag,
			TaggedAt: d.Pricing.TaggedAt.Format(time.RFC3339),
			Currency: string(d.Pricing.Currency),
		}
	}

	if d.SaleInfo != nil {
		resp.Selling = &VehicleSellingSection{
			SalePrice:   d.SaleInfo.SalePrice,
			SaleDate:    d.SaleInfo.SaleDate.Format(time.RFC3339),
			PaymentMode: d.SaleInfo.PaymentMode,
			ReceiptUrl:  d.SaleInfo.ReceiptUrl,
			Remarks:     d.SaleInfo.Remarks,
			Customer: VehicleSaleCustomer{
				FirstName: d.SaleInfo.CustomerFirstName,
				LastName:  d.SaleInfo.CustomerLastName,
				Email:     d.SaleInfo.CustomerEmail,
				Phone:     d.SaleInfo.CustomerPhone,
				Address:   d.SaleInfo.CustomerAddress,
				City:      d.SaleInfo.CustomerCity,
				State:     d.SaleInfo.CustomerState,
			},
		}
	}

	return resp
}

func buildBasicResponse(d *VehicleFullDetails) GetVehicleBasicResponse {
	resp := GetVehicleBasicResponse{
		Basic: buildBasicSection(d),
	}

	if d.Pricing != nil {
		resp.Pricing = &VehiclePriceTagOnly{
			PriceTag: d.Pricing.PriceTag,
			Currency: string(d.Pricing.Currency),
		}
	}

	if d.SaleInfo != nil {
		resp.Selling = &VehicleSoldPriceOnly{
			SalePrice: d.SaleInfo.SalePrice,
		}
	}

	return resp
}

func buildBasicSection(d *VehicleFullDetails) VehicleBasicSection {
	section := VehicleBasicSection{
		ID:                 d.Vehicle.ID,
		VehicleType:        string(d.Vehicle.VehicleType),
		Manufacturer:       d.Vehicle.Manufacturer,
		Model:              d.Vehicle.Model,
		Variant:            d.Vehicle.Variant,
		Color:              d.Vehicle.Color,
		YearOfManufacture:  d.Vehicle.YearOfManufacture,
		RTOCode:            d.Vehicle.RTOCode,
		RegistrationNumber: d.Vehicle.RegistrationNumber,
		RegistrationState:  d.Vehicle.RegistrationState,
		UsageKM:            d.Vehicle.UsageKM,
		FuelType:           string(d.Vehicle.FuelType),
		TransmissionType:   string(d.Vehicle.TransmissionType),
		CreatedAt:          d.Vehicle.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          d.Vehicle.UpdatedAt.Format(time.RFC3339),
	}

	if len(d.Statuses) > 0 {
		section.CurrentStatus = &VehicleStatusSummary{
			Status:    string(d.Statuses[0].Status),
			StartedAt: d.Statuses[0].StartedAt.Format(time.RFC3339),
		}
	}

	return section
}

func buildStatusSection(statuses []VehicleStatus) VehicleStatusSection {
	items := make([]VehicleStatusItem, 0, len(statuses))
	for _, s := range statuses {
		items = append(items, VehicleStatusItem{
			Status:      string(s.Status),
			Description: s.Description,
			StartedAt:   s.StartedAt.Format(time.RFC3339),
			EndedAt:     s.EndedAt.Format(time.RFC3339),
		})
	}

	section := VehicleStatusSection{History: items}
	if len(items) > 0 {
		section.Current = &items[0]
	}
	return section
}

func buildExpenseItems(expenses []VehicleExpenses) []VehicleExpenseItem {
	items := make([]VehicleExpenseItem, 0, len(expenses))
	for _, e := range expenses {
		items = append(items, VehicleExpenseItem{
			ID:          e.ID,
			Type:        string(e.Type),
			Amount:      e.Amount,
			PaidTo:      e.PaidTo,
			Description: e.Description,
			Date:        e.Date.Format(time.RFC3339),
		})
	}
	return items
}

func buildDocumentItems(docs []VehicleDocument) []VehicleDocumentItem {
	items := make([]VehicleDocumentItem, 0, len(docs))
	for _, d := range docs {
		items = append(items, VehicleDocumentItem{
			ID:           d.ID,
			DocumentType: string(d.DocumentType),
			DocumentURL:  d.DocumentURL,
			ValidFrom:    d.ValidFrom.Format(time.RFC3339),
			ValidTill:    d.ValidTill.Format(time.RFC3339),
			Remarks:      d.Remarks,
			UploadedAt:   d.UploadedAt.Format(time.RFC3339),
		})
	}
	return items
}

func (h *Handler) UpdateVehicle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	var req UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	showroomRolesVal, exists := c.Get(middleware.ContextKeyShowroomRoles)
	if !exists {
		response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
		return
	}
	showroomRoles, ok := showroomRolesVal.(map[uint64]string)
	if !ok {
		response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
		return
	}

	showroomID, err := h.service.GetVehicleShowroomID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	if _, isMember := showroomRoles[showroomID]; !isMember {
		response.Error(c, http.StatusNotFound, apperrors.CodeVehicleNotFound, "vehicle not found")
		return
	}

	resp, err := h.service.UpdateVehicle(c.Request.Context(), id, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "vehicle updated", resp)
}

func (h *Handler) UpdateVehiclePricing(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	var req UpdateVehiclePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidRequest, "invalid request")
		return
	}

	showroomRolesVal, exists := c.Get(middleware.ContextKeyShowroomRoles)
	if !exists {
		response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
		return
	}
	showroomRoles, ok := showroomRolesVal.(map[uint64]string)
	if !ok {
		response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
		return
	}

	showroomID, err := h.service.GetVehicleShowroomID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	if _, isMember := showroomRoles[showroomID]; !isMember {
		response.Error(c, http.StatusNotFound, apperrors.CodeVehicleNotFound, "vehicle not found")
		return
	}

	resp, err := h.service.UpdateVehiclePricing(c.Request.Context(), id, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, "vehicle pricing updated", resp)
}

func buildImageItems(images []VehicleImage) []VehicleImageItem {
	items := make([]VehicleImageItem, 0, len(images))
	for _, img := range images {
		items = append(items, VehicleImageItem{
			ID:    img.ID,
			Label: string(img.Label),
			URL:   img.ImageURL,
		})
	}
	return items
}
