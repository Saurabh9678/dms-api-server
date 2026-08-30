package vehicle

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"infiour.local/dms-api-server/pkg/database"
	"infiour.local/dms-api-server/pkg/inventory"
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
	SoldBy            *uint64
	SoldByName        string
	SoldByCountryCode string
	SoldByPhoneNumber string
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
	SoldBy            *uint64   `gorm:"column:sold_by"`
	SoldByName        string    `gorm:"column:sold_by_name"`
	SoldByCountryCode string    `gorm:"column:sold_by_country_code"`
	SoldByPhoneNumber string    `gorm:"column:sold_by_phone_number"`
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
		       cvs.sold_by, cvs.sold_by_name, cvs.sold_by_country_code, cvs.sold_by_phone_number,
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
			SoldBy:            sale.SoldBy,
			SoldByName:        sale.SoldByName,
			SoldByCountryCode: sale.SoldByCountryCode,
			SoldByPhoneNumber: sale.SoldByPhoneNumber,
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
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(vehicle).Error; err != nil {
			return err
		}
		status := VehicleStatus{
			VehicleID: vehicle.ID,
			Status:    VehicleStatusTypeGarage,
			StartedAt: time.Now(),
		}
		return tx.Select("VehicleID", "Status", "StartedAt").Create(&status).Error
	})
	if err != nil {
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

// saleCustomerRecord persists a per-sale customer snapshot without importing the customer module.
type saleCustomerRecord struct {
	ID          uint64 `gorm:"column:id;primaryKey"`
	FirstName   string `gorm:"column:first_name"`
	LastName    string `gorm:"column:last_name"`
	Email       string `gorm:"column:email"`
	PhoneNumber string `gorm:"column:phone_number"`
	Address     string `gorm:"column:address"`
	City        string `gorm:"column:city"`
	State       string `gorm:"column:state"`
	Pincode     string `gorm:"column:pincode"`
	database.SoftDeleteableModel
}

func (saleCustomerRecord) TableName() string { return "customers" }

type saleRecord struct {
	ID                uint64    `gorm:"column:id;primaryKey"`
	CustomerID        uint64    `gorm:"column:customer_id"`
	VehicleID         uint64    `gorm:"column:vehicle_id"`
	SalePrice         float64   `gorm:"column:sale_price"`
	SaleDate          time.Time `gorm:"column:sale_date"`
	PaymentMode       string    `gorm:"column:payment_mode"`
	ReceiptUrl        string    `gorm:"column:receipt_url"`
	Remarks           string    `gorm:"column:remarks"`
	SoldBy            uint64    `gorm:"column:sold_by"`
	SoldByName        string    `gorm:"column:sold_by_name"`
	SoldByCountryCode string    `gorm:"column:sold_by_country_code"`
	SoldByPhoneNumber string    `gorm:"column:sold_by_phone_number"`
	database.SoftDeleteableModel
}

func (saleRecord) TableName() string { return "customer_vehicle_sales" }

type sellerUserRecord struct {
	ID          uint64 `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	CountryCode string `gorm:"column:country_code"`
	PhoneNumber string `gorm:"column:phone_number"`
	database.SoftDeleteableModel
}

func (sellerUserRecord) TableName() string { return "users" }

// SellVehicleInput is the persistence payload for recording a vehicle sale.
type SellVehicleInput struct {
	VehicleID   uint64
	SoldBy      uint64
	SalePrice   float64
	SaleDate    time.Time
	PaymentMode string
	Remarks     string
	FirstName   string
	LastName    string
	PhoneNumber string
	Address     string
	Email       string
	City        string
	State       string
	Pincode     string
}

// SellVehicleResult is returned after a successful sale transaction.
type SellVehicleResult struct {
	SaleID            uint64
	CustomerID        uint64
	VehicleID         uint64
	SalePrice         float64
	SaleDate          time.Time
	PaymentMode       string
	Remarks           string
	FirstName         string
	LastName          string
	PhoneNumber       string
	Address           string
	Email             string
	City              string
	State             string
	Pincode           string
	SoldBy            uint64
	SoldByName        string
	SoldByCountryCode string
	SoldByPhoneNumber string
}

// SellVehicle creates a customer snapshot, sale row, and sold status in one transaction.
func (r *Repository) SellVehicle(ctx context.Context, in SellVehicleInput) (*SellVehicleResult, error) {
	var result SellVehicleResult
	err := database.RunInTx(ctx, r.db, func(tx *gorm.DB) error {
		var statusRow struct {
			Status VehicleStatusType `gorm:"column:status"`
		}
		statusQuery := tx.WithContext(ctx).Raw(
			"SELECT status FROM vehicle_statuses WHERE vehicle_id = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT 1",
			in.VehicleID,
		).Scan(&statusRow)
		if statusQuery.Error != nil {
			return statusQuery.Error
		}
		if statusQuery.RowsAffected == 0 {
			return ErrVehicleNotFound
		}
		if statusRow.Status == VehicleStatusTypeSold {
			return ErrVehicleAlreadySold
		}

		var saleCount int64
		if err := tx.WithContext(ctx).
			Table("customer_vehicle_sales").
			Where("vehicle_id = ? AND deleted_at IS NULL", in.VehicleID).
			Count(&saleCount).Error; err != nil {
			return err
		}
		if saleCount > 0 {
			return ErrVehicleAlreadySold
		}

		var seller sellerUserRecord
		if err := tx.WithContext(ctx).First(&seller, in.SoldBy).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrVehicleNotFound // should not happen for authenticated callers; treat as not found
			}
			return err
		}

		customer := saleCustomerRecord{
			FirstName:   in.FirstName,
			LastName:    in.LastName,
			Email:       in.Email,
			PhoneNumber: in.PhoneNumber,
			Address:     in.Address,
			City:        in.City,
			State:       in.State,
			Pincode:     in.Pincode,
		}
		if err := tx.WithContext(ctx).Create(&customer).Error; err != nil {
			return err
		}

		sale := saleRecord{
			CustomerID:        customer.ID,
			VehicleID:         in.VehicleID,
			SalePrice:         in.SalePrice,
			SaleDate:          in.SaleDate,
			PaymentMode:       in.PaymentMode,
			Remarks:           in.Remarks,
			SoldBy:            in.SoldBy,
			SoldByName:        seller.Name,
			SoldByCountryCode: seller.CountryCode,
			SoldByPhoneNumber: seller.PhoneNumber,
		}
		if err := tx.WithContext(ctx).Create(&sale).Error; err != nil {
			return err
		}

		status := VehicleStatus{
			VehicleID: in.VehicleID,
			Status:    VehicleStatusTypeSold,
			StartedAt: time.Now().UTC(),
			AddedBy:   in.SoldBy,
		}
		if err := tx.WithContext(ctx).Select("VehicleID", "Status", "StartedAt", "AddedBy").Create(&status).Error; err != nil {
			return err
		}

		result = SellVehicleResult{
			SaleID:            sale.ID,
			CustomerID:        customer.ID,
			VehicleID:         in.VehicleID,
			SalePrice:         in.SalePrice,
			SaleDate:          in.SaleDate,
			PaymentMode:       in.PaymentMode,
			Remarks:           in.Remarks,
			FirstName:         in.FirstName,
			LastName:          in.LastName,
			PhoneNumber:       in.PhoneNumber,
			Address:           in.Address,
			Email:             in.Email,
			City:              in.City,
			State:             in.State,
			Pincode:           in.Pincode,
			SoldBy:            in.SoldBy,
			SoldByName:        seller.Name,
			SoldByCountryCode: seller.CountryCode,
			SoldByPhoneNumber: seller.PhoneNumber,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateStatus closes the current status row (ended_at = now) and inserts the new status.
func (r *Repository) UpdateStatus(ctx context.Context, vehicleID, addedBy uint64, status VehicleStatusType, description string) (*VehicleStatus, error) {
	var created VehicleStatus
	err := database.RunInTx(ctx, r.db, func(tx *gorm.DB) error {
		var current struct {
			ID     uint64            `gorm:"column:id"`
			Status VehicleStatusType `gorm:"column:status"`
		}
		q := tx.WithContext(ctx).Raw(
			"SELECT id, status FROM vehicle_statuses WHERE vehicle_id = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT 1",
			vehicleID,
		).Scan(&current)
		if q.Error != nil {
			return q.Error
		}
		if q.RowsAffected == 0 {
			return ErrVehicleNotFound
		}
		if current.Status == VehicleStatusTypeSold {
			return ErrVehicleSold
		}
		if current.Status == status {
			return ErrVehicleStatusUnchanged
		}

		now := time.Now().UTC()
		if err := tx.WithContext(ctx).
			Model(&VehicleStatus{}).
			Where("id = ?", current.ID).
			Update("ended_at", now).Error; err != nil {
			return err
		}

		created = VehicleStatus{
			VehicleID:   vehicleID,
			Status:      status,
			Description: description,
			StartedAt:   now,
			AddedBy:     addedBy,
		}
		return tx.WithContext(ctx).
			Select("VehicleID", "Status", "Description", "StartedAt", "AddedBy").
			Create(&created).Error
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *Repository) CreateImage(ctx context.Context, img *VehicleImage) (*VehicleImage, error) {
	if err := r.db.WithContext(ctx).Create(img).Error; err != nil {
		return nil, err
	}
	return img, nil
}

func (r *Repository) SoftDeleteImage(ctx context.Context, vehicleID, imageID uint64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND vehicle_id = ?", imageID, vehicleID).Delete(&VehicleImage{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrVehicleImageNotFound
	}
	return nil
}

// ListImagesByVehicleIDs returns non-deleted images keyed by vehicle_id.
func (r *Repository) ListImagesByVehicleIDs(ctx context.Context, vehicleIDs []uint64) (map[uint64][]VehicleImage, error) {
	out := make(map[uint64][]VehicleImage)
	if len(vehicleIDs) == 0 {
		return out, nil
	}
	var images []VehicleImage
	if err := r.db.WithContext(ctx).Where("vehicle_id IN ?", vehicleIDs).Order("id ASC").Find(&images).Error; err != nil {
		return nil, err
	}
	for _, img := range images {
		out[img.VehicleID] = append(out[img.VehicleID], img)
	}
	return out, nil
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
	ShowroomID   uint64
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
	VpBuyingDate       *time.Time       `gorm:"column:vp_buying_date"`
	HasActiveSale      bool             `gorm:"column:has_active_sale"`
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
	Images             []VehicleImage
	BuyingDate         *time.Time
	HasActiveSale      bool
}

type CategoryMetrics struct {
	Total          int64
	DeadStockCount int64
}

// StatusBreakdownCounts are unpaginated per-type status tallies (status filter ignored).
type StatusBreakdownCounts struct {
	AvailableCount int64
	RepairCount    int64
	SoldCount      int64
}

func buildListQuery(filter ListFilter) (string, []interface{}) {
	query := `
SELECT v.id, v.type AS vehicle_type, v.manufacturer, v.model, v.variant, v.color,
       v.year_of_manufacture, v.rto_code, v.registration_number, v.registration_state,
       v.usage_km, v.fuel_type, v.transmission_type, v.created_at, v.updated_at,
       vs.status AS vs_status, vs.started_at AS vs_started_at,
       vp.buying_price AS vp_buying_price, vp.price_tag AS vp_price_tag,
       vp.currency AS vp_currency, vp.tagged_at AS vp_tagged_at,
       vp.buying_date AS vp_buying_date,
       EXISTS (
         SELECT 1 FROM customer_vehicle_sales cvs
         WHERE cvs.vehicle_id = v.id AND cvs.deleted_at IS NULL
       ) AS has_active_sale
FROM vehicles v
JOIN vehicle_showroom_relations vsr ON vsr.vehicle_id = v.id
  AND vsr.showroom_id = ?
  AND vsr.deleted_at IS NULL
JOIN LATERAL (
  SELECT status, started_at FROM vehicle_statuses
  WHERE vehicle_id = v.id AND deleted_at IS NULL
  ORDER BY id DESC LIMIT 1
) vs ON true
LEFT JOIN LATERAL (
  SELECT buying_price, price_tag, currency, tagged_at, buying_date FROM vehicle_pricing
  WHERE vehicle_id = v.id AND deleted_at IS NULL
  ORDER BY id DESC LIMIT 1
) vp ON true
WHERE v.deleted_at IS NULL
  AND (? OR vs.status IN (?))
  AND (? OR v.type IN (?))
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

	noStatusFilter := len(statuses) == 0
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
		filter.ShowroomID,
		noStatusFilter, statuses,
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
			BuyingDate:         row.VpBuyingDate,
			HasActiveSale:      row.HasActiveSale,
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
			if row.VpBuyingDate != nil {
				p.BuyingDate = *row.VpBuyingDate
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
SELECT v.id, v.type AS vehicle_type, v.manufacturer, v.model, v.variant, v.color,
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
  AND (? OR v.type IN (?))
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
	VehicleType    VehicleType `gorm:"column:vehicle_type"`
	Count          int64       `gorm:"column:count"`
	DeadStockCount int64       `gorm:"column:dead_stock_count"`
}

func (r *Repository) CountByType(ctx context.Context, filter ListFilter) (map[VehicleType]CategoryMetrics, error) {
	query, args := buildListQuery(filter)
	countQuery := `
SELECT vq.vehicle_type,
       COUNT(*) AS count,
       COALESCE(SUM(CASE WHEN NOT vq.has_active_sale AND ` + inventory.DeadStockAgePredicateSQL("vq.vp_buying_date") + ` THEN 1 ELSE 0 END), 0) AS dead_stock_count
FROM (` + query + `) vq
GROUP BY vq.vehicle_type`

	var rows []vehicleTypeCount
	if err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[VehicleType]CategoryMetrics)
	for _, row := range rows {
		result[row.VehicleType] = CategoryMetrics{
			Total:          row.Count,
			DeadStockCount: row.DeadStockCount,
		}
	}
	return result, nil
}

type vehicleStatusBreakdownRow struct {
	VehicleType    VehicleType `gorm:"column:vehicle_type"`
	AvailableCount int64       `gorm:"column:available_count"`
	RepairCount    int64       `gorm:"column:repair_count"`
	SoldCount      int64       `gorm:"column:sold_count"`
}

// CountStatusBreakdownByType returns available/repair/sold counts per type.
// Status query filters are ignored so tallies always cover all current statuses
// (still scoped by showroom, type, and price filters).
func (r *Repository) CountStatusBreakdownByType(ctx context.Context, filter ListFilter) (map[VehicleType]StatusBreakdownCounts, error) {
	filter.Statuses = nil
	query, args := buildListQuery(filter)
	countQuery := `
SELECT vq.vehicle_type,
       COALESCE(SUM(CASE WHEN vq.vs_status = 'ready_for_sale' THEN 1 ELSE 0 END), 0) AS available_count,
       COALESCE(SUM(CASE WHEN vq.vs_status IN ('garage', 'inspection') THEN 1 ELSE 0 END), 0) AS repair_count,
       COALESCE(SUM(CASE WHEN vq.vs_status = 'sold' THEN 1 ELSE 0 END), 0) AS sold_count
FROM (` + query + `) vq
GROUP BY vq.vehicle_type`

	var rows []vehicleStatusBreakdownRow
	if err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[VehicleType]StatusBreakdownCounts)
	for _, row := range rows {
		result[row.VehicleType] = StatusBreakdownCounts{
			AvailableCount: row.AvailableCount,
			RepairCount:    row.RepairCount,
			SoldCount:      row.SoldCount,
		}
	}
	return result, nil
}
