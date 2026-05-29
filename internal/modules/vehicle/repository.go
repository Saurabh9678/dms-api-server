package vehicle

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, vehicle *Vehicle) (*Vehicle, error) {
	if err := r.db.WithContext(ctx).Create(vehicle).Error; err != nil {
		return nil, err
	}
	return vehicle, nil
}

type ListFilter struct {
	Statuses     []VehicleStatusType
	VehicleTypes []VehicleType
	MinPrice     *float64
	MaxPrice     *float64
	Page         int
	Limit        int
}

type vehicleRow struct {
	ID                 uint64           `gorm:"column:id"`
	VehicleType        VehicleType      `gorm:"column:vehicle_type"`
	Manufacturer       string           `gorm:"column:manufacturer"`
	Model              string           `gorm:"column:model"`
	Variant            string           `gorm:"column:variant"`
	Color              string           `gorm:"column:color"`
	YearOfManufacture  int              `gorm:"column:year_of_manufacture"`
	RTOCode            string           `gorm:"column:rto_code"`
	RegistrationNumber string           `gorm:"column:registration_number"`
	RegistrationState  string           `gorm:"column:registration_state"`
	UsageKM            int              `gorm:"column:usage_km"`
	FuelType           FuelType         `gorm:"column:fuel_type"`
	TransmissionType   TransmissionType `gorm:"column:transmission_type"`
	CreatedAt          time.Time        `gorm:"column:created_at"`
	UpdatedAt          time.Time        `gorm:"column:updated_at"`
	VsStatus           *string          `gorm:"column:vs_status"`
	VsStartedAt        *time.Time       `gorm:"column:vs_started_at"`
	VpBuyingPrice      *float64         `gorm:"column:vp_buying_price"`
	VpPriceTag         *float64         `gorm:"column:vp_price_tag"`
	VpCurrency         *string          `gorm:"column:vp_currency"`
	VpTaggedAt         *time.Time       `gorm:"column:vp_tagged_at"`
}

type VehicleWithDetails struct {
	ID                 uint64
	VehicleType        VehicleType
	Manufacturer       string
	Model              string
	Variant            string
	Color              string
	YearOfManufacture  int
	RTOCode            string
	RegistrationNumber string
	RegistrationState  string
	UsageKM            int
	FuelType           FuelType
	TransmissionType   TransmissionType
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CurrentStatus      *VehicleStatus
	CurrentPricing     *VehiclePricing
}

func buildListQuery(filter ListFilter) (string, []interface{}) {
	query := `
SELECT v.id, v.vehicle_type, v.manufacturer, v.model, v.variant, v.color,
       v.year_of_manufacture, v.rto_code, v.registration_number, v.registration_state,
       v.usage_km, v.fuel_type, v.transmission_type, v.created_at, v.updated_at,
       vs.status AS vs_status, vs.started_at AS vs_started_at,
       vp.buying_price AS vp_buying_price, vp.price_tag AS vp_price_tag,
       vp.currency AS vp_currency, vp.tagged_at AS vp_tagged_at
FROM vehicles v
JOIN LATERAL (
  SELECT status, started_at FROM vehicle_statuses
  WHERE vehicle_id = v.id AND deleted_at IS NULL
  ORDER BY id DESC LIMIT 1
) vs ON true
LEFT JOIN LATERAL (
  SELECT buying_price, price_tag, currency, tagged_at FROM vehicle_pricing
  WHERE vehicle_id = v.id AND deleted_at IS NULL
  ORDER BY id DESC LIMIT 1
) vp ON true
WHERE v.deleted_at IS NULL
  AND vs.status = ANY(?)
  AND (? OR v.vehicle_type = ANY(?))
  AND (? OR vp.price_tag >= ?)
  AND (? OR vp.price_tag <= ?)
ORDER BY v.id DESC`

	statuses := make([]string, len(filter.Statuses))
	for i, s := range filter.Statuses {
		statuses[i] = string(s)
	}

	types := make([]string, len(filter.VehicleTypes))
	for i, t := range filter.VehicleTypes {
		types[i] = string(t)
	}

	noTypeFilter := len(types) == 0
	noMinPrice := filter.MinPrice == nil
	noMaxPrice := filter.MaxPrice == nil

	minPrice := 0.0
	if filter.MinPrice != nil {
		minPrice = *filter.MinPrice
	}
	maxPrice := 0.0
	if filter.MaxPrice != nil {
		maxPrice = *filter.MaxPrice
	}

	args := []interface{}{
		statuses,
		noTypeFilter, types,
		noMinPrice, minPrice,
		noMaxPrice, maxPrice,
	}
	return query, args
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]VehicleWithDetails, error) {
	query, args := buildListQuery(filter)
	query += "\nLIMIT ? OFFSET ?"
	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var rows []vehicleRow
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	results := make([]VehicleWithDetails, 0, len(rows))
	for _, row := range rows {
		v := VehicleWithDetails{
			ID:                 row.ID,
			VehicleType:        row.VehicleType,
			Manufacturer:       row.Manufacturer,
			Model:              row.Model,
			Variant:            row.Variant,
			Color:              row.Color,
			YearOfManufacture:  row.YearOfManufacture,
			RTOCode:            row.RTOCode,
			RegistrationNumber: row.RegistrationNumber,
			RegistrationState:  row.RegistrationState,
			UsageKM:            row.UsageKM,
			FuelType:           row.FuelType,
			TransmissionType:   row.TransmissionType,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
		}
		if row.VsStatus != nil {
			st := VehicleStatus{Status: VehicleStatusType(*row.VsStatus)}
			if row.VsStartedAt != nil {
				st.StartedAt = *row.VsStartedAt
			}
			v.CurrentStatus = &st
		}
		if row.VpPriceTag != nil {
			p := VehiclePricing{BuyingPrice: 0, PriceTag: *row.VpPriceTag}
			if row.VpBuyingPrice != nil {
				p.BuyingPrice = *row.VpBuyingPrice
			}
			if row.VpCurrency != nil {
				p.Currency = Currency(*row.VpCurrency)
			}
			if row.VpTaggedAt != nil {
				p.TaggedAt = *row.VpTaggedAt
			}
			v.CurrentPricing = &p
		}
		results = append(results, v)
	}
	return results, nil
}

type vehicleTypeCount struct {
	VehicleType VehicleType `gorm:"column:vehicle_type"`
	Count       int64       `gorm:"column:count"`
}

func (r *Repository) CountByType(ctx context.Context, filter ListFilter) (map[VehicleType]int64, error) {
	query, args := buildListQuery(filter)
	countQuery := "SELECT vq.vehicle_type, COUNT(*) AS count FROM (" + query + ") vq GROUP BY vq.vehicle_type"

	var rows []vehicleTypeCount
	if err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[VehicleType]int64)
	for _, row := range rows {
		result[row.VehicleType] = row.Count
	}
	return result, nil
}
