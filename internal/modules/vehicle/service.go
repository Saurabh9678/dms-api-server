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
}

type vehicleRepo interface {
	Create(ctx context.Context, vehicle *Vehicle) (*Vehicle, error)
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

func isValidVehicleType(vt VehicleType) bool {
	return vt == VehicleTypeBike || vt == VehicleTypeCar || vt == VehicleTypeScooty
}

func isValidFuelType(ft FuelType) bool {
	return ft == FuelTypePetrol || ft == FuelTypeDiesel || ft == FuelTypeEV
}

func isValidTransmissionType(tt TransmissionType) bool {
	return tt == TransmissionTypeManual || tt == TransmissionTypeAutomatic
}
