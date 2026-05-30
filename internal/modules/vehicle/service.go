package vehicle

import (
	"context"
	"net/http"
	"strings"
	"time"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

type Service interface {
	CreateVehicle(ctx context.Context, req *CreateVehicleRequest, addedBy uint64) (*CreateVehicleResponse, error)
	ListVehicles(ctx context.Context, query *ListVehiclesQuery) (*ListVehiclesResponse, error)
	GetVehicleByID(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error)
	PublicListVehicles(ctx context.Context, query *PublicListVehiclesQuery) (*PublicListVehiclesResponse, error)
	GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error)
	UpdateVehicle(ctx context.Context, vehicleID uint64, req *UpdateVehicleRequest) (*UpdateVehicleResponse, error)
	UpdateVehiclePricing(ctx context.Context, vehicleID uint64, req *UpdateVehiclePricingRequest) (*UpdateVehiclePricingResponse, error)
}

type vehicleRepo interface {
	CreateWithInitialStatus(ctx context.Context, vehicle *Vehicle, status *VehicleStatus) (*Vehicle, error)
	List(ctx context.Context, f ListFilter) ([]VehicleWithDetails, error)
	CountByType(ctx context.Context, f ListFilter) (map[VehicleType]int64, error)
	GetByIDWithFullDetails(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error)
	PublicList(ctx context.Context, f PublicListFilter) ([]VehicleWithDetails, error)
	PublicCountByType(ctx context.Context, f PublicListFilter) (map[VehicleType]int64, error)
	GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error)
	GetCurrentStatus(ctx context.Context, vehicleID uint64) (VehicleStatusType, error)
	UpdateVehicleFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*Vehicle, error)
	GetPricingByVehicleID(ctx context.Context, vehicleID uint64) (*VehiclePricing, error)
	CreatePricing(ctx context.Context, pricing *VehiclePricing) (*VehiclePricing, error)
	UpdatePricingFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*VehiclePricing, error)
}

type service struct {
	repo vehicleRepo
}

func NewService(repo vehicleRepo) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateVehicle(ctx context.Context, req *CreateVehicleRequest, addedBy uint64) (*CreateVehicleResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	vehicle := &Vehicle{
		VehicleType:        req.VehicleType,
		Manufacturer:       strings.TrimSpace(req.Manufacturer),
		Model:              strings.TrimSpace(req.Model),
		Variant:            strings.TrimSpace(req.Variant),
		Color:              strings.TrimSpace(req.Color),
		YearOfManufacture:  req.YearOfManufacture,
		RTOCode:            strings.TrimSpace(req.RTOCode),
		RegistrationNumber: strings.TrimSpace(req.RegistrationNumber),
		RegistrationState:  strings.TrimSpace(req.RegistrationState),
		UsageKM:            req.UsageKM,
		FuelType:           req.FuelType,
		TransmissionType:   req.TransmissionType,
	}

	now := time.Now()
	status := &VehicleStatus{
		Status:    VehicleStatusTypeBought,
		StartedAt: now,
		EndedAt:   now,
		AddedBy:   addedBy,
	}

	created, err := s.repo.CreateWithInitialStatus(ctx, vehicle, status)
	if err != nil {
		return nil, err
	}

	return &CreateVehicleResponse{
		ID:                 created.ID,
		VehicleType:        string(created.VehicleType),
		Manufacturer:       created.Manufacturer,
		Model:              created.Model,
		Variant:            created.Variant,
		Color:              created.Color,
		YearOfManufacture:  created.YearOfManufacture,
		RTOCode:            created.RTOCode,
		RegistrationNumber: created.RegistrationNumber,
		RegistrationState:  created.RegistrationState,
		UsageKM:            created.UsageKM,
		FuelType:           string(created.FuelType),
		TransmissionType:   string(created.TransmissionType),
		CreatedAt:          created.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          created.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) validateRequest(req *CreateVehicleRequest) error {
	if req == nil {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if !isValidVehicleType(req.VehicleType) {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if strings.TrimSpace(req.Manufacturer) == "" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if strings.TrimSpace(req.Model) == "" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if strings.TrimSpace(req.Variant) == "" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if strings.TrimSpace(req.Color) == "" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	currentYear := time.Now().Year()
	if req.YearOfManufacture < 1900 || req.YearOfManufacture > currentYear {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if strings.TrimSpace(req.RTOCode) == "" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if strings.TrimSpace(req.RegistrationNumber) == "" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if strings.TrimSpace(req.RegistrationState) == "" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if req.UsageKM < 0 {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if !isValidFuelType(req.FuelType) {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if !isValidTransmissionType(req.TransmissionType) {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	return nil
}

func (s *service) ListVehicles(ctx context.Context, query *ListVehiclesQuery) (*ListVehiclesResponse, error) {
	if err := s.validateListQuery(query); err != nil {
		return nil, err
	}

	statuses := make([]VehicleStatusType, 0, len(query.Statuses))
	if len(query.Statuses) == 0 {
		statuses = append(statuses, VehicleStatusTypeReadyForSale)
	} else {
		for _, s := range query.Statuses {
			statuses = append(statuses, VehicleStatusType(s))
		}
	}

	types := make([]VehicleType, 0, len(query.VehicleTypes))
	for _, t := range query.VehicleTypes {
		types = append(types, VehicleType(t))
	}

	filter := ListFilter{
		Statuses:     statuses,
		VehicleTypes: types,
		MinPrice:     query.MinPrice,
		MaxPrice:     query.MaxPrice,
		Page:         query.Page,
		Limit:        query.Limit,
	}

	counts, err := s.repo.CountByType(ctx, filter)
	if err != nil {
		return nil, err
	}

	vehicles, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	grouped := map[VehicleType][]VehicleListItem{
		VehicleTypeCar:    {},
		VehicleTypeBike:   {},
		VehicleTypeScooty: {},
	}
	for _, v := range vehicles {
		item := toVehicleListItem(v)
		grouped[v.VehicleType] = append(grouped[v.VehicleType], item)
	}

	wantType := map[VehicleType]bool{}
	if len(types) == 0 {
		wantType[VehicleTypeCar] = true
		wantType[VehicleTypeBike] = true
		wantType[VehicleTypeScooty] = true
	} else {
		for _, t := range types {
			wantType[t] = true
		}
	}

	resp := &ListVehiclesResponse{}
	if wantType[VehicleTypeCar] {
		resp.Cars = &CategoryListing{
			Total:    counts[VehicleTypeCar],
			Page:     query.Page,
			Limit:    query.Limit,
			Vehicles: grouped[VehicleTypeCar],
		}
	}
	if wantType[VehicleTypeBike] {
		resp.Bikes = &CategoryListing{
			Total:    counts[VehicleTypeBike],
			Page:     query.Page,
			Limit:    query.Limit,
			Vehicles: grouped[VehicleTypeBike],
		}
	}
	if wantType[VehicleTypeScooty] {
		resp.Scooties = &CategoryListing{
			Total:    counts[VehicleTypeScooty],
			Page:     query.Page,
			Limit:    query.Limit,
			Vehicles: grouped[VehicleTypeScooty],
		}
	}

	return resp, nil
}

func (s *service) PublicListVehicles(ctx context.Context, query *PublicListVehiclesQuery) (*PublicListVehiclesResponse, error) {
	if err := s.validatePublicListQuery(query); err != nil {
		return nil, err
	}

	types := make([]VehicleType, 0, len(query.VehicleTypes))
	for _, t := range query.VehicleTypes {
		types = append(types, VehicleType(t))
	}

	filter := PublicListFilter{
		ShowroomID:   query.ShowroomID,
		VehicleTypes: types,
		MinPrice:     query.MinPrice,
		MaxPrice:     query.MaxPrice,
		SortBy:       query.SortBy,
		Page:         query.Page,
		Limit:        query.Limit,
	}

	counts, err := s.repo.PublicCountByType(ctx, filter)
	if err != nil {
		return nil, err
	}

	vehicles, err := s.repo.PublicList(ctx, filter)
	if err != nil {
		return nil, err
	}

	grouped := map[VehicleType][]PublicVehicleListItem{
		VehicleTypeCar:    {},
		VehicleTypeBike:   {},
		VehicleTypeScooty: {},
	}
	for _, v := range vehicles {
		item := toPublicVehicleListItem(v)
		grouped[v.VehicleType] = append(grouped[v.VehicleType], item)
	}

	wantType := map[VehicleType]bool{}
	if len(types) == 0 {
		wantType[VehicleTypeCar] = true
		wantType[VehicleTypeBike] = true
		wantType[VehicleTypeScooty] = true
	} else {
		for _, t := range types {
			wantType[t] = true
		}
	}

	resp := &PublicListVehiclesResponse{}
	if wantType[VehicleTypeCar] {
		resp.Cars = &PublicCategoryListing{
			Total:    counts[VehicleTypeCar],
			Page:     query.Page,
			Limit:    query.Limit,
			Vehicles: grouped[VehicleTypeCar],
		}
	}
	if wantType[VehicleTypeBike] {
		resp.Bikes = &PublicCategoryListing{
			Total:    counts[VehicleTypeBike],
			Page:     query.Page,
			Limit:    query.Limit,
			Vehicles: grouped[VehicleTypeBike],
		}
	}
	if wantType[VehicleTypeScooty] {
		resp.Scooties = &PublicCategoryListing{
			Total:    counts[VehicleTypeScooty],
			Page:     query.Page,
			Limit:    query.Limit,
			Vehicles: grouped[VehicleTypeScooty],
		}
	}

	return resp, nil
}

func (s *service) validatePublicListQuery(query *PublicListVehiclesQuery) error {
	if query == nil {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if query.ShowroomID == 0 {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if query.Page < 1 {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if query.Limit < 1 || query.Limit > 100 {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if query.SortBy != "price_asc" && query.SortBy != "price_desc" {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	for _, t := range query.VehicleTypes {
		if !isValidVehicleType(VehicleType(t)) {
			return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
	}
	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	return nil
}

func toPublicVehicleListItem(v VehicleWithDetails) PublicVehicleListItem {
	item := PublicVehicleListItem{
		ID:                 v.ID,
		VehicleType:        string(v.VehicleType),
		Manufacturer:       v.Manufacturer,
		Model:              v.Model,
		Variant:            v.Variant,
		Color:              v.Color,
		YearOfManufacture:  v.YearOfManufacture,
		RTOCode:            v.RTOCode,
		RegistrationNumber: v.RegistrationNumber,
		RegistrationState:  v.RegistrationState,
		UsageKM:            v.UsageKM,
		FuelType:           string(v.FuelType),
		TransmissionType:   string(v.TransmissionType),
		CreatedAt:          v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          v.UpdatedAt.Format(time.RFC3339),
	}
	if v.CurrentPricing != nil {
		item.PriceTag = v.CurrentPricing.PriceTag
		item.Currency = string(v.CurrentPricing.Currency)
	}
	return item
}

func toVehicleListItem(v VehicleWithDetails) VehicleListItem {
	item := VehicleListItem{
		ID:                 v.ID,
		VehicleType:        string(v.VehicleType),
		Manufacturer:       v.Manufacturer,
		Model:              v.Model,
		Variant:            v.Variant,
		Color:              v.Color,
		YearOfManufacture:  v.YearOfManufacture,
		RTOCode:            v.RTOCode,
		RegistrationNumber: v.RegistrationNumber,
		RegistrationState:  v.RegistrationState,
		UsageKM:            v.UsageKM,
		FuelType:           string(v.FuelType),
		TransmissionType:   string(v.TransmissionType),
		CreatedAt:          v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          v.UpdatedAt.Format(time.RFC3339),
	}
	if v.CurrentStatus != nil {
		item.CurrentStatus = &VehicleStatusSummary{
			Status:    string(v.CurrentStatus.Status),
			StartedAt: v.CurrentStatus.StartedAt.Format(time.RFC3339),
		}
	}
	if v.CurrentPricing != nil {
		item.Pricing = &VehiclePricingSummary{
			BuyingPrice: v.CurrentPricing.BuyingPrice,
			PriceTag:    v.CurrentPricing.PriceTag,
			Currency:    string(v.CurrentPricing.Currency),
			TaggedAt:    v.CurrentPricing.TaggedAt.Format(time.RFC3339),
		}
	}
	return item
}

func (s *service) validateListQuery(query *ListVehiclesQuery) error {
	if query == nil {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if query.Page < 1 {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if query.Limit < 1 || query.Limit > 100 {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	for _, s := range query.Statuses {
		if !isValidVehicleStatusType(VehicleStatusType(s)) {
			return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
	}
	for _, t := range query.VehicleTypes {
		if !isValidVehicleType(VehicleType(t)) {
			return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
	}
	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	return nil
}

func (s *service) GetVehicleByID(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error) {
	return s.repo.GetByIDWithFullDetails(ctx, vehicleID)
}

func isValidVehicleStatusType(st VehicleStatusType) bool {
	return st == VehicleStatusTypeGarage ||
		st == VehicleStatusTypeInspection ||
		st == VehicleStatusTypeReadyForSale ||
		st == VehicleStatusTypeSold
}

func isValidVehicleType(vt VehicleType) bool {
	return vt == VehicleTypeBike || vt == VehicleTypeCar || vt == VehicleTypeScooty
}

func isValidFuelType(ft FuelType) bool {
	return ft == FuelTypePetrol || ft == FuelTypeDiesel || ft == FuelTypeEV
}

func isValidTransmissionType(tt TransmissionType) bool {
	return tt == TransmissionTypeManual || tt == TransmissionTypeAutomatic
}

func isValidCurrency(c Currency) bool {
	return c == CurrencyINR || c == CurrencyUSD
}

func (s *service) GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error) {
	return s.repo.GetVehicleShowroomID(ctx, vehicleID)
}

func (s *service) UpdateVehicle(ctx context.Context, vehicleID uint64, req *UpdateVehicleRequest) (*UpdateVehicleResponse, error) {
	if req == nil {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	status, err := s.repo.GetCurrentStatus(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if status == VehicleStatusTypeSold {
		return nil, ErrVehicleSold
	}

	updates, err := s.buildVehicleUpdates(req)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	updated, err := s.repo.UpdateVehicleFields(ctx, vehicleID, updates)
	if err != nil {
		return nil, err
	}

	return &UpdateVehicleResponse{
		ID:                 updated.ID,
		VehicleType:        string(updated.VehicleType),
		Manufacturer:       updated.Manufacturer,
		Model:              updated.Model,
		Variant:            updated.Variant,
		Color:              updated.Color,
		YearOfManufacture:  updated.YearOfManufacture,
		RTOCode:            updated.RTOCode,
		RegistrationNumber: updated.RegistrationNumber,
		RegistrationState:  updated.RegistrationState,
		UsageKM:            updated.UsageKM,
		FuelType:           string(updated.FuelType),
		TransmissionType:   string(updated.TransmissionType),
		UpdatedAt:          updated.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) buildVehicleUpdates(req *UpdateVehicleRequest) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if req.VehicleType != nil {
		if !isValidVehicleType(*req.VehicleType) {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["vehicle_type"] = *req.VehicleType
	}
	if req.Manufacturer != nil {
		trimmed := strings.TrimSpace(*req.Manufacturer)
		if trimmed == "" {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["manufacturer"] = trimmed
	}
	if req.Model != nil {
		trimmed := strings.TrimSpace(*req.Model)
		if trimmed == "" {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["model"] = trimmed
	}
	if req.Variant != nil {
		trimmed := strings.TrimSpace(*req.Variant)
		if trimmed == "" {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["variant"] = trimmed
	}
	if req.Color != nil {
		trimmed := strings.TrimSpace(*req.Color)
		if trimmed == "" {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["color"] = trimmed
	}
	if req.YearOfManufacture != nil {
		currentYear := time.Now().Year()
		if *req.YearOfManufacture < 1900 || *req.YearOfManufacture > currentYear {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["year_of_manufacture"] = *req.YearOfManufacture
	}
	if req.RTOCode != nil {
		trimmed := strings.TrimSpace(*req.RTOCode)
		if trimmed == "" {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["rto_code"] = trimmed
	}
	if req.RegistrationState != nil {
		trimmed := strings.TrimSpace(*req.RegistrationState)
		if trimmed == "" {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["registration_state"] = trimmed
	}
	if req.UsageKM != nil {
		if *req.UsageKM < 0 {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["usage_km"] = *req.UsageKM
	}
	if req.FuelType != nil {
		if !isValidFuelType(*req.FuelType) {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["fuel_type"] = *req.FuelType
	}
	if req.TransmissionType != nil {
		if !isValidTransmissionType(*req.TransmissionType) {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["transmission_type"] = *req.TransmissionType
	}

	return updates, nil
}

func (s *service) UpdateVehiclePricing(ctx context.Context, vehicleID uint64, req *UpdateVehiclePricingRequest) (*UpdateVehiclePricingResponse, error) {
	if req == nil {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	status, err := s.repo.GetCurrentStatus(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if status == VehicleStatusTypeSold {
		return nil, ErrVehicleSold
	}

	existing, err := s.repo.GetPricingByVehicleID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return s.createPricing(ctx, vehicleID, req)
	}

	updates, err := s.buildPricingUpdates(req)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	updated, err := s.repo.UpdatePricingFields(ctx, vehicleID, updates)
	if err != nil {
		return nil, err
	}

	return toPricingResponse(updated), nil
}

func (s *service) createPricing(ctx context.Context, vehicleID uint64, req *UpdateVehiclePricingRequest) (*UpdateVehiclePricingResponse, error) {
	if req.BuyingPrice == nil || *req.BuyingPrice <= 0 {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if req.BuyingDate == nil {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	buyingDate, err := time.Parse("2006-01-02", *req.BuyingDate)
	if err != nil {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	var priceTag float64
	if req.PriceTag != nil {
		if *req.PriceTag < 0 {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		priceTag = *req.PriceTag
	}

	taggedAt := time.Now()
	if req.TaggedAt != nil {
		parsed, parseErr := time.Parse(time.RFC3339, *req.TaggedAt)
		if parseErr != nil {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		taggedAt = parsed
	}

	currency := CurrencyINR
	if req.Currency != nil {
		if !isValidCurrency(Currency(*req.Currency)) {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		currency = Currency(*req.Currency)
	}

	remarks := ""
	if req.Remarks != nil {
		remarks = *req.Remarks
	}

	newPricing := &VehiclePricing{
		VehicleID:   vehicleID,
		BuyingPrice: *req.BuyingPrice,
		BuyingDate:  buyingDate,
		PriceTag:    priceTag,
		TaggedAt:    taggedAt,
		Currency:    currency,
		Remarks:     remarks,
	}

	created, err := s.repo.CreatePricing(ctx, newPricing)
	if err != nil {
		return nil, err
	}

	return toPricingResponse(created), nil
}

func (s *service) buildPricingUpdates(req *UpdateVehiclePricingRequest) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	if req.BuyingPrice != nil {
		if *req.BuyingPrice <= 0 {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["buying_price"] = *req.BuyingPrice
	}
	if req.BuyingDate != nil {
		date, parseErr := time.Parse("2006-01-02", *req.BuyingDate)
		if parseErr != nil {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["buying_date"] = date
	}
	if req.PriceTag != nil {
		if *req.PriceTag < 0 {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["price_tag"] = *req.PriceTag
	}
	if req.TaggedAt != nil {
		taggedAt, parseErr := time.Parse(time.RFC3339, *req.TaggedAt)
		if parseErr != nil {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["tagged_at"] = taggedAt
	}
	if req.Currency != nil {
		if !isValidCurrency(Currency(*req.Currency)) {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		updates["currency"] = *req.Currency
	}
	if req.Remarks != nil {
		updates["remarks"] = *req.Remarks
	}

	return updates, nil
}

func toPricingResponse(p *VehiclePricing) *UpdateVehiclePricingResponse {
	return &UpdateVehiclePricingResponse{
		VehicleID:   p.VehicleID,
		BuyingPrice: p.BuyingPrice,
		BuyingDate:  p.BuyingDate.Format("2006-01-02"),
		PriceTag:    p.PriceTag,
		TaggedAt:    p.TaggedAt.Format(time.RFC3339),
		Currency:    string(p.Currency),
		Remarks:     p.Remarks,
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}
