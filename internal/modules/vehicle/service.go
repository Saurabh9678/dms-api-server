package vehicle

import (
	"context"
	"net/http"
	"strings"
	"time"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

type Service interface {
	CreateVehicle(ctx context.Context, req *CreateVehicleRequest) (*CreateVehicleResponse, error)
	ListVehicles(ctx context.Context, query *ListVehiclesQuery) (*ListVehiclesResponse, error)
	GetVehicleByID(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error)
}

type vehicleRepo interface {
	Create(ctx context.Context, vehicle *Vehicle) (*Vehicle, error)
	List(ctx context.Context, f ListFilter) ([]VehicleWithDetails, error)
	CountByType(ctx context.Context, f ListFilter) (map[VehicleType]int64, error)
	GetByIDWithFullDetails(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error)
}

type service struct {
	repo vehicleRepo
}

func NewService(repo vehicleRepo) Service {
	return &service{
		repo: repo,
	}
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
