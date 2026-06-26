package vehicle

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type VehicleSaleInfo struct {
	SalePrice         float64
	SaleDate          time.Time
	PaymentMode       string
	ReceiptUrl        string
	Remarks           string
	CustomerFirstName string
	CustomerLastName  string
	CustomerEmail     string
	CustomerPhone     string
	CustomerAddress   string
	CustomerCity      string
	CustomerState     string
}

type VehicleFullDetails struct {
	Vehicle    Vehicle
	Pricing    *VehiclePricing
	Statuses   []VehicleStatus
	Documents  []VehicleDocument
	Expenses   []VehicleExpenses
	Images     []VehicleImage
	ShowroomID uint64
	SaleInfo   *VehicleSaleInfo
}

type saleRow struct {
	SalePrice         float64   `gorm:"column:sale_price"`
	SaleDate          time.Time `gorm:"column:sale_date"`
	PaymentMode       string    `gorm:"column:payment_mode"`
	ReceiptUrl        string    `gorm:"column:receipt_url"`
	Remarks           string    `gorm:"column:remarks"`
	CustomerFirstName string    `gorm:"column:customer_first_name"`
	CustomerLastName  string    `gorm:"column:customer_last_name"`
	CustomerEmail     string    `gorm:"column:customer_email"`
	CustomerPhone     string    `gorm:"column:customer_phone"`
	CustomerAddress   string    `gorm:"column:customer_address"`
	CustomerCity      string    `gorm:"column:customer_city"`
	CustomerState     string    `gorm:"column:customer_state"`
}

func (r *Repository) GetByIDWithFullDetails(ctx context.Context, vehicleID uint64) (*VehicleFullDetails, error) {
	var v Vehicle
	if err := r.db.WithContext(ctx).First(&v, vehicleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVehicleNotFound
		}
		return nil, err
	}

	details := &VehicleFullDetails{Vehicle: v}

	var pricing VehiclePricing
	if err := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Order("id DESC").First(&pricing).Error; err == nil {
		details.Pricing = &pricing
	}

	var statuses []VehicleStatus
	r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Order("started_at DESC").Find(&statuses)
	details.Statuses = statuses

	var docs []VehicleDocument
	r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Find(&docs)
	details.Documents = docs

	var expenses []VehicleExpenses
	r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Find(&expenses)
	details.Expenses = expenses

	var images []VehicleImage
	r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).Find(&images)
	details.Images = images

	var showroomRel VehicleShowroom
	if err := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).First(&showroomRel).Error; err == nil {
		details.ShowroomID = showroomRel.ShowroomID
	}

	var sale saleRow
	result := r.db.WithContext(ctx).Raw(`
		SELECT cvs.sale_price, cvs.sale_date, cvs.payment_mode, cvs.receipt_url, cvs.remarks,
		       c.first_name AS customer_first_name, c.last_name AS customer_last_name,
		       c.email AS customer_email, c.phone_number AS customer_phone,
		       c.address AS customer_address, c.city AS customer_city, c.state AS customer_state
		FROM customer_vehicle_sales cvs
		JOIN customers c ON c.id = cvs.customer_id
		WHERE cvs.vehicle_id = ? AND cvs.deleted_at IS NULL
		ORDER BY cvs.id DESC LIMIT 1`, vehicleID).Scan(&sale)
	if result.Error == nil && result.RowsAffected > 0 {
		details.SaleInfo = &VehicleSaleInfo{
			SalePrice:         sale.SalePrice,
			SaleDate:          sale.SaleDate,
			PaymentMode:       sale.PaymentMode,
			ReceiptUrl:        sale.ReceiptUrl,
			Remarks:           sale.Remarks,
			CustomerFirstName: sale.CustomerFirstName,
			CustomerLastName:  sale.CustomerLastName,
			CustomerEmail:     sale.CustomerEmail,
			CustomerPhone:     sale.CustomerPhone,
			CustomerAddress:   sale.CustomerAddress,
			CustomerCity:      sale.CustomerCity,
			CustomerState:     sale.CustomerState,
		}
	}

	return details, nil
}

func (r *Repository) GetVehicleShowroomID(ctx context.Context, vehicleID uint64) (uint64, error) {
	var rel VehicleShowroom
	err := r.db.WithContext(ctx).Where("vehicle_id = ? AND deleted_at IS NULL", vehicleID).First(&rel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrVehicleNotFound
	}
	if err != nil {
		return 0, err
	}
	return rel.ShowroomID, nil
}

func (r *Repository) GetCurrentStatus(ctx context.Context, vehicleID uint64) (VehicleStatusType, error) {
	var row struct {
		Status VehicleStatusType `gorm:"column:status"`
	}
	result := r.db.WithContext(ctx).Raw(
		"SELECT status FROM vehicle_statuses WHERE vehicle_id = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT 1",
		vehicleID,
	).Scan(&row)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", ErrVehicleNotFound
	}
	return row.Status, nil
}

func (r *Repository) UpdateVehicleFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*Vehicle, error) {
	result := r.db.WithContext(ctx).Model(&Vehicle{}).Where("id = ?", vehicleID).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrVehicleNotFound
	}
	var updated Vehicle
	if err := r.db.WithContext(ctx).First(&updated, vehicleID).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *Repository) GetPricingByVehicleID(ctx context.Context, vehicleID uint64) (*VehiclePricing, error) {
	var pricing VehiclePricing
	err := r.db.WithContext(ctx).Where("vehicle_id = ? AND deleted_at IS NULL", vehicleID).First(&pricing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pricing, nil
}

func (r *Repository) CreatePricing(ctx context.Context, pricing *VehiclePricing) (*VehiclePricing, error) {
	if err := r.db.WithContext(ctx).Create(pricing).Error; err != nil {
		return nil, err
	}
	return pricing, nil
}

func (r *Repository) UpdatePricingFields(ctx context.Context, vehicleID uint64, updates map[string]interface{}) (*VehiclePricing, error) {
	result := r.db.WithContext(ctx).Model(&VehiclePricing{}).Where("vehicle_id = ?", vehicleID).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	var updated VehiclePricing
	if err := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).First(&updated).Error; err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *Repository) Create(ctx context.Context, vehicle *Vehicle) (*Vehicle, error) {
	if err := r.db.WithContext(ctx).Create(vehicle).Error; err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (r *Repository) CreateExpense(ctx context.Context, expense *VehicleExpenses) (*VehicleExpenses, error) {
	if err := r.db.WithContext(ctx).Create(expense).Error; err != nil {
		return nil, err
	}
	return expense, nil
}

func (r *Repository) VehicleExistsByID(ctx context.Context, vehicleID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Vehicle{}).Where("id = ? AND deleted_at IS NULL", vehicleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) AssignShowroom(ctx context.Context, vehicleID, showroomID uint64) (*VehicleShowroom, error) {
	rel := &VehicleShowroom{
		VehicleID:  vehicleID,
		ShowroomID: showroomID,
	}
	if err := r.db.WithContext(ctx).Create(rel).Error; err != nil {
		return nil, err
	}
	return rel, nil
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

type PublicListFilter struct {
	ShowroomID   uint64
	VehicleTypes []VehicleType
	MinPrice     *float64
	MaxPrice     *float64
	SortBy       string
	Page         int
	Limit        int
}

func buildPublicListQuery(filter PublicListFilter) (string, []interface{}) {
	orderClause := "vp.price_tag ASC"
	if filter.SortBy == "price_desc" {
		orderClause = "vp.price_tag DESC"
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

	query := `
SELECT v.id, v.vehicle_type, v.manufacturer, v.model, v.variant, v.color,
       v.year_of_manufacture, v.rto_code, v.registration_number, v.registration_state,
       v.usage_km, v.fuel_type, v.transmission_type, v.created_at, v.updated_at,
       vs.status AS vs_status, vs.started_at AS vs_started_at,
       vp.buying_price AS vp_buying_price, vp.price_tag AS vp_price_tag,
       vp.currency AS vp_currency, vp.tagged_at AS vp_tagged_at
FROM vehicles v
JOIN vehicle_showroom_relations vsr ON vsr.vehicle_id = v.id
  AND vsr.showroom_id = ?
  AND vsr.deleted_at IS NULL
JOIN LATERAL (
  SELECT status, started_at FROM vehicle_statuses
  WHERE vehicle_id = v.id AND deleted_at IS NULL
  ORDER BY id DESC LIMIT 1
) vs ON true
JOIN LATERAL (
  SELECT buying_price, price_tag, currency, tagged_at FROM vehicle_pricing
  WHERE vehicle_id = v.id AND deleted_at IS NULL AND price_tag IS NOT NULL
  ORDER BY id DESC LIMIT 1
) vp ON true
WHERE v.deleted_at IS NULL
  AND vs.status = 'ready_for_sale'
  AND (? OR v.vehicle_type = ANY(?))
  AND (? OR vp.price_tag >= ?)
  AND (? OR vp.price_tag <= ?)
ORDER BY ` + orderClause

	args := []interface{}{
		filter.ShowroomID,
		noTypeFilter, types,
		noMinPrice, minPrice,
		noMaxPrice, maxPrice,
	}
	return query, args
}

func (r *Repository) PublicList(ctx context.Context, filter PublicListFilter) ([]VehicleWithDetails, error) {
	query, args := buildPublicListQuery(filter)
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
			p := VehiclePricing{PriceTag: *row.VpPriceTag}
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

func (r *Repository) PublicCountByType(ctx context.Context, filter PublicListFilter) (map[VehicleType]int64, error) {
	query, args := buildPublicListQuery(filter)
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
