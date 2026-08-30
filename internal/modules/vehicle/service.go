package vehicle

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	storageprovider "infiour.local/dms-api-server/internal/providers/storage"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/inventory"
)

const maxVehicleImageSize = 15 * 1024 * 1024 // 15 MB

var allowedVehicleImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// ServiceOption configures the vehicle service. Used in tests to inject behaviour.
type ServiceOption func(*service)

// WithFileOpener overrides how multipart file headers are opened. Used in tests.
func WithFileOpener(fn func(*multipart.FileHeader) (io.ReadCloser, error)) ServiceOption {
	return func(s *service) { s.openFile = fn }
}

// WithSignedURLTTL overrides the signed URL lifetime used in API responses.
func WithSignedURLTTL(ttl time.Duration) ServiceOption {
	return func(s *service) {
		if ttl > 0 {
			s.signedURLTTL = ttl
		}
	}
}

type Service interface {
	CreateVehicle(ctx context.Context, req *CreateVehicleRequest) (*CreateVehicleResponse, error)
	ListVehicles(ctx context.Context, query *ListVehiclesQuery) (*ListVehiclesResponse, error)
	GetVehicleByID(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error)
	PublicListVehicles(ctx context.Context, query *PublicListVehiclesQuery) (*PublicListVehiclesResponse, error)
	GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error)
	UpdateVehicle(ctx context.Context, vehicleID uint64, req *UpdateVehicleRequest) (*UpdateVehicleResponse, error)
	UpdateVehiclePricing(ctx context.Context, vehicleID uint64, req *UpdateVehiclePricingRequest) (*UpdateVehiclePricingResponse, error)
	AddExpense(ctx context.Context, vehicleID uint64, req *AddExpenseRequest) (*AddExpenseResponse, error)
	SellVehicle(ctx context.Context, sellerUserID, vehicleID uint64, req *SellVehicleRequest) (*SellVehicleResponse, error)
	AssignVehicleToShowroom(ctx context.Context, vehicleID, showroomID uint64) (*AssignShowroomResponse, error)
	AddVehicleImage(ctx context.Context, userID, vehicleID uint64, label string, photo *multipart.FileHeader) (*AddVehicleImageResponse, error)
	DeleteVehicleImage(ctx context.Context, vehicleID, imageID uint64) error
}

type vehicleRepo interface {
	Create(ctx context.Context, vehicle *Vehicle) (*Vehicle, error)
	List(ctx context.Context, f ListFilter) ([]VehicleWithDetails, error)
	CountByType(ctx context.Context, f ListFilter) (map[VehicleType]CategoryMetrics, error)
	CountStatusBreakdownByType(ctx context.Context, f ListFilter) (map[VehicleType]StatusBreakdownCounts, error)
	GetByIDWithFullDetails(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error)
	PublicList(ctx context.Context, f PublicListFilter) ([]VehicleWithDetails, error)
	PublicCountByType(ctx context.Context, f PublicListFilter) (map[VehicleType]int64, error)
	GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error)
	GetCurrentStatus(ctx context.Context, vehicleID uint64) (VehicleStatusType, error)
	UpdateVehicleFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*Vehicle, error)
	GetPricingByVehicleID(ctx context.Context, vehicleID uint64) (*VehiclePricing, error)
	CreatePricing(ctx context.Context, pricing *VehiclePricing) (*VehiclePricing, error)
	UpdatePricingFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*VehiclePricing, error)
	CreateExpense(ctx context.Context, expense *VehicleExpenses) (*VehicleExpenses, error)
	SellVehicle(ctx context.Context, in SellVehicleInput) (*SellVehicleResult, error)
	VehicleExistsByID(ctx context.Context, vehicleID uint64) (bool, error)
	AssignShowroom(ctx context.Context, vehicleID, showroomID uint64) (*VehicleShowroom, error)
	CreateImage(ctx context.Context, img *VehicleImage) (*VehicleImage, error)
	SoftDeleteImage(ctx context.Context, vehicleID, imageID uint64) error
	ListImagesByVehicleIDs(ctx context.Context, vehicleIDs []uint64) (map[uint64][]VehicleImage, error)
}

type service struct {
	repo         vehicleRepo
	storage      storageprovider.Provider
	signedURLTTL time.Duration
	openFile     func(*multipart.FileHeader) (io.ReadCloser, error)
}

func NewService(repo vehicleRepo, storage storageprovider.Provider, opts ...ServiceOption) Service {
	s := &service{
		repo:         repo,
		storage:      storage,
		signedURLTTL: time.Hour,
		openFile: func(h *multipart.FileHeader) (io.ReadCloser, error) {
			return h.Open()
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *service) CreateVehicle(ctx context.Context, req *CreateVehicleRequest) (*CreateVehicleResponse, error) {
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

	created, err := s.repo.Create(ctx, vehicle)
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
	for _, s := range query.Statuses {
		statuses = append(statuses, VehicleStatusType(s))
	}

	types := make([]VehicleType, 0, len(query.VehicleTypes))
	for _, t := range query.VehicleTypes {
		types = append(types, VehicleType(t))
	}

	filter := ListFilter{
		ShowroomID:   query.ShowroomID,
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

	statusCounts, err := s.repo.CountStatusBreakdownByType(ctx, filter)
	if err != nil {
		return nil, err
	}

	vehicles, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err := s.attachListImages(ctx, vehicles); err != nil {
		return nil, err
	}

	grouped := map[VehicleType][]VehicleListItem{
		VehicleTypeCar:    {},
		VehicleTypeBike:   {},
		VehicleTypeScooty: {},
	}
	for _, v := range vehicles {
		item := s.toVehicleListItem(ctx, v)
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

	buildCategory := func(t VehicleType) *CategoryListing {
		return &CategoryListing{
			Total:          counts[t].Total,
			AvailableCount: statusCounts[t].AvailableCount,
			RepairCount:    statusCounts[t].RepairCount,
			SoldCount:      statusCounts[t].SoldCount,
			DeadStockCount: counts[t].DeadStockCount,
			Page:           query.Page,
			Limit:          query.Limit,
			Vehicles:       grouped[t],
		}
	}

	resp := &ListVehiclesResponse{}
	if wantType[VehicleTypeCar] {
		resp.Cars = buildCategory(VehicleTypeCar)
	}
	if wantType[VehicleTypeBike] {
		resp.Bikes = buildCategory(VehicleTypeBike)
	}
	if wantType[VehicleTypeScooty] {
		resp.Scooties = buildCategory(VehicleTypeScooty)
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

	if err := s.attachListImages(ctx, vehicles); err != nil {
		return nil, err
	}

	grouped := map[VehicleType][]PublicVehicleListItem{
		VehicleTypeCar:    {},
		VehicleTypeBike:   {},
		VehicleTypeScooty: {},
	}
	for _, v := range vehicles {
		item := s.toPublicVehicleListItem(ctx, v)
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

func (s *service) attachListImages(ctx context.Context, vehicles []VehicleWithDetails) error {
	if len(vehicles) == 0 {
		return nil
	}
	ids := make([]uint64, len(vehicles))
	for i := range vehicles {
		ids[i] = vehicles[i].ID
	}
	byVehicle, err := s.repo.ListImagesByVehicleIDs(ctx, ids)
	if err != nil {
		return err
	}
	for i := range vehicles {
		imgs := byVehicle[vehicles[i].ID]
		if imgs == nil {
			imgs = []VehicleImage{}
		}
		vehicles[i].Images = imgs
	}
	return nil
}

func (s *service) toPublicVehicleListItem(ctx context.Context, v VehicleWithDetails) PublicVehicleListItem {
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
		Images:             s.buildSignedImageSection(ctx, v.Images),
		CreatedAt:          v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          v.UpdatedAt.Format(time.RFC3339),
	}
	if v.CurrentPricing != nil {
		item.PriceTag = v.CurrentPricing.PriceTag
		item.Currency = string(v.CurrentPricing.Currency)
	}
	return item
}

func (s *service) toVehicleListItem(ctx context.Context, v VehicleWithDetails) VehicleListItem {
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
		Images:             s.buildSignedImageSection(ctx, v.Images),
		IsDeadStock:        inventory.IsDeadStock(v.BuyingDate, v.HasActiveSale, time.Now()),
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
	if query.ShowroomID == 0 {
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
	details, err := s.repo.GetByIDWithFullDetails(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	s.signVehicleImages(ctx, details)
	return details, nil
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
		updates["type"] = *req.VehicleType
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

func (s *service) AddExpense(ctx context.Context, vehicleID uint64, req *AddExpenseRequest) (*AddExpenseResponse, error) {
	if req == nil {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if !isValidExpenseType(VehicleExpensesType(req.Type)) {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	if req.Amount <= 0 {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	date := time.Now()
	if req.Date != nil {
		parsed, err := time.Parse(time.RFC3339, *req.Date)
		if err != nil {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		date = parsed
	}

	expense := &VehicleExpenses{
		VehicleID:   vehicleID,
		Type:        VehicleExpensesType(req.Type),
		Amount:      req.Amount,
		PaidTo:      strings.TrimSpace(req.PaidTo),
		Description: strings.TrimSpace(req.Description),
		Date:        date,
	}

	created, err := s.repo.CreateExpense(ctx, expense)
	if err != nil {
		return nil, err
	}

	return &AddExpenseResponse{
		ID:          created.ID,
		VehicleID:   created.VehicleID,
		Type:        string(created.Type),
		Amount:      created.Amount,
		PaidTo:      created.PaidTo,
		Description: created.Description,
		Date:        created.Date.Format(time.RFC3339),
		CreatedAt:   created.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) SellVehicle(ctx context.Context, sellerUserID, vehicleID uint64, req *SellVehicleRequest) (*SellVehicleResponse, error) {
	if req == nil {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if sellerUserID == 0 {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidAccessToken, "invalid request", http.StatusUnauthorized, nil)
	}

	if req.SalePrice <= 0 {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if !isValidPaymentMode(req.PaymentMode) {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	firstName := strings.TrimSpace(req.Customer.FirstName)
	lastName := strings.TrimSpace(req.Customer.LastName)
	phoneNumber := strings.TrimSpace(req.Customer.PhoneNumber)
	address := strings.TrimSpace(req.Customer.Address)
	if firstName == "" || lastName == "" || phoneNumber == "" || address == "" {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}

	saleDate := time.Now().UTC()
	saleDate = time.Date(saleDate.Year(), saleDate.Month(), saleDate.Day(), 0, 0, 0, 0, time.UTC)
	if req.SaleDate != nil && strings.TrimSpace(*req.SaleDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*req.SaleDate))
		if err != nil {
			return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
		}
		saleDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	}

	remarks := ""
	if req.Remarks != nil {
		remarks = strings.TrimSpace(*req.Remarks)
	}

	in := SellVehicleInput{
		VehicleID:   vehicleID,
		SoldBy:      sellerUserID,
		SalePrice:   req.SalePrice,
		SaleDate:    saleDate,
		PaymentMode: req.PaymentMode,
		Remarks:     remarks,
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phoneNumber,
		Address:     address,
		Email:       optionalTrimmed(req.Customer.Email),
		City:        optionalTrimmed(req.Customer.City),
		State:       optionalTrimmed(req.Customer.State),
		Pincode:     optionalTrimmed(req.Customer.Pincode),
	}

	created, err := s.repo.SellVehicle(ctx, in)
	if err != nil {
		return nil, err
	}

	resp := &SellVehicleResponse{
		ID:          created.SaleID,
		VehicleID:   created.VehicleID,
		SalePrice:   created.SalePrice,
		SaleDate:    created.SaleDate.Format("2006-01-02"),
		PaymentMode: created.PaymentMode,
		Customer: SellVehicleCustomerResponse{
			ID:          created.CustomerID,
			FirstName:   created.FirstName,
			LastName:    created.LastName,
			PhoneNumber: created.PhoneNumber,
			Address:     created.Address,
			Email:       optionalStringPtr(created.Email),
			City:        optionalStringPtr(created.City),
			State:       optionalStringPtr(created.State),
			Pincode:     optionalStringPtr(created.Pincode),
		},
		SoldBy: VehicleSoldBy{
			UserID:      created.SoldBy,
			Name:        optionalStringPtr(created.SoldByName),
			CountryCode: optionalStringPtr(created.SoldByCountryCode),
			PhoneNumber: optionalStringPtr(created.SoldByPhoneNumber),
		},
	}
	if created.Remarks != "" {
		r := created.Remarks
		resp.Remarks = &r
	}
	return resp, nil
}

func isValidPaymentMode(mode string) bool {
	switch mode {
	case "cash", "cheque", "bank_transfer", "online", "credit", "debit", "other":
		return true
	default:
		return false
	}
}

func optionalTrimmed(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func optionalStringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func isValidExpenseType(t VehicleExpensesType) bool {
	return t == VehicleExpensesTypeRepair ||
		t == VehicleExpensesTypeService ||
		t == VehicleExpensesTypeInsurance ||
		t == VehicleExpensesTypeTax ||
		t == VehicleExpensesTypeInspection ||
		t == VehicleExpensesTypeCleaning ||
		t == VehicleExpensesTypeDocumentation ||
		t == VehicleExpensesTypeOther
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

func (s *service) AssignVehicleToShowroom(ctx context.Context, vehicleID, showroomID uint64) (*AssignShowroomResponse, error) {
	exists, err := s.repo.VehicleExistsByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrVehicleNotFound
	}

	_, err = s.repo.GetVehicleShowroomID(ctx, vehicleID)
	if err == nil {
		return nil, ErrVehicleAlreadyInShowroom
	}
	if !stderrors.Is(err, ErrVehicleNotFound) {
		return nil, err
	}

	rel, err := s.repo.AssignShowroom(ctx, vehicleID, showroomID)
	if err != nil {
		return nil, err
	}

	return &AssignShowroomResponse{
		VehicleID:  rel.VehicleID,
		ShowroomID: rel.ShowroomID,
		AssignedAt: rel.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) AddVehicleImage(ctx context.Context, userID, vehicleID uint64, label string, photo *multipart.FileHeader) (*AddVehicleImageResponse, error) {
	normalizedLabel := strings.TrimSpace(strings.ToLower(label))
	if !isValidVehicleImageLabel(VehicleImageLabel(normalizedLabel)) {
		return nil, apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if err := validateVehicleImageFile(photo); err != nil {
		return nil, err
	}

	status, err := s.repo.GetCurrentStatus(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if status == VehicleStatusTypeSold {
		return nil, ErrVehicleSold
	}

	key, err := s.readAndUploadImage(ctx, userID, vehicleID, photo)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.CreateImage(ctx, &VehicleImage{
		VehicleID:  vehicleID,
		ImageURL:   key,
		Label:      VehicleImageLabel(normalizedLabel),
		UploadedAt: time.Now(),
		UploadedBy: userID,
	})
	if err != nil {
		return nil, err
	}

	url := ""
	if signed, signErr := s.storage.SignedURL(ctx, created.ImageURL, s.signedURLTTL); signErr == nil {
		url = signed
	}

	return &AddVehicleImageResponse{
		ID:         created.ID,
		VehicleID:  created.VehicleID,
		Label:      string(created.Label),
		URL:        url,
		UploadedAt: created.UploadedAt.Format(time.RFC3339),
	}, nil
}

func (s *service) DeleteVehicleImage(ctx context.Context, vehicleID, imageID uint64) error {
	status, err := s.repo.GetCurrentStatus(ctx, vehicleID)
	if err != nil {
		return err
	}
	if status == VehicleStatusTypeSold {
		return ErrVehicleSold
	}
	return s.repo.SoftDeleteImage(ctx, vehicleID, imageID)
}

func (s *service) readAndUploadImage(ctx context.Context, userID, vehicleID uint64, header *multipart.FileHeader) (string, error) {
	f, err := s.openFile(header)
	if err != nil {
		return "", apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, err)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	key := fmt.Sprintf("%d/vehicle/%d/%s%s", userID, vehicleID, time.Now().Format("20060102150405"), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	storedKey, err := s.storage.Upload(ctx, key, data, contentType)
	if err != nil {
		return "", apperrors.NewAppError(apperrors.CodeInternal, "failed to upload image", http.StatusInternalServerError, err)
	}
	return storedKey, nil
}

func (s *service) signVehicleImages(ctx context.Context, details *VehicleFullDetails) {
	if details == nil || len(details.Images) == 0 {
		return
	}
	kept := make([]VehicleImage, 0, len(details.Images))
	for _, img := range details.Images {
		if img.ImageURL == "" {
			continue
		}
		url, err := s.storage.SignedURL(ctx, img.ImageURL, s.signedURLTTL)
		if err != nil {
			continue
		}
		img.ImageURL = url
		kept = append(kept, img)
	}
	details.Images = kept
}

// buildSignedImageSection groups images by label with signed URLs.
// Always returns a non-nil map (empty object in JSON when there are no photos).
func (s *service) buildSignedImageSection(ctx context.Context, images []VehicleImage) map[string][]VehicleImageItem {
	section := make(map[string][]VehicleImageItem)
	for _, img := range images {
		label := string(img.Label)
		if label == "" || img.ImageURL == "" {
			continue
		}
		url, err := s.storage.SignedURL(ctx, img.ImageURL, s.signedURLTTL)
		if err != nil {
			continue
		}
		section[label] = append(section[label], VehicleImageItem{
			ID:  img.ID,
			URL: url,
		})
	}
	return section
}

func validateVehicleImageFile(header *multipart.FileHeader) error {
	if header == nil {
		return apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
	}
	if header.Size > maxVehicleImageSize {
		return apperrors.NewAppError(apperrors.CodeFileTooLarge, "invalid request", http.StatusBadRequest, nil)
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedVehicleImageExtensions[ext] {
		return apperrors.NewAppError(apperrors.CodeInvalidFileType, "invalid request", http.StatusBadRequest, nil)
	}
	return nil
}

func isValidVehicleImageLabel(label VehicleImageLabel) bool {
	return label == VehicleImageLabelFront ||
		label == VehicleImageLabelInterior ||
		label == VehicleImageLabelExterior ||
		label == VehicleImageLabelBack ||
		label == VehicleImageLabelWheel
}
