package vehicle

type ListVehiclesQuery struct {
	Statuses     []string `form:"status"`
	VehicleTypes []string `form:"type"`
	MinPrice     *float64 `form:"min_price"`
	MaxPrice     *float64 `form:"max_price"`
	Page         int      `form:"page,default=1"`
	Limit        int      `form:"limit,default=20"`
}

type VehicleStatusSummary struct {
	Status    string `json:"status"`
	StartedAt string `json:"started_at"`
}

type VehiclePricingSummary struct {
	BuyingPrice float64 `json:"buying_price"`
	PriceTag    float64 `json:"price_tag"`
	Currency    string  `json:"currency"`
	TaggedAt    string  `json:"tagged_at"`
}

type VehicleListItem struct {
	ID                 uint64                 `json:"id"`
	VehicleType        string                 `json:"vehicle_type"`
	Manufacturer       string                 `json:"manufacturer"`
	Model              string                 `json:"model"`
	Variant            string                 `json:"variant"`
	Color              string                 `json:"color"`
	YearOfManufacture  int                    `json:"year_of_manufacture"`
	RTOCode            string                 `json:"rto_code"`
	RegistrationNumber string                 `json:"registration_number"`
	RegistrationState  string                 `json:"registration_state"`
	UsageKM            int                    `json:"usage_km"`
	FuelType           string                 `json:"fuel_type"`
	TransmissionType   string                 `json:"transmission_type"`
	CurrentStatus      *VehicleStatusSummary  `json:"current_status"`
	Pricing            *VehiclePricingSummary `json:"pricing"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

type CategoryListing struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Vehicles []VehicleListItem `json:"vehicles"`
}

type ListVehiclesResponse struct {
	Cars     *CategoryListing `json:"cars,omitempty"`
	Bikes    *CategoryListing `json:"bikes,omitempty"`
	Scooties *CategoryListing `json:"scooties,omitempty"`
}

type VehicleBasicSection struct {
	ID                 uint64                `json:"id"`
	VehicleType        string                `json:"vehicle_type"`
	Manufacturer       string                `json:"manufacturer"`
	Model              string                `json:"model"`
	Variant            string                `json:"variant"`
	Color              string                `json:"color"`
	YearOfManufacture  int                   `json:"year_of_manufacture"`
	RTOCode            string                `json:"rto_code"`
	RegistrationNumber string                `json:"registration_number"`
	RegistrationState  string                `json:"registration_state"`
	UsageKM            int                   `json:"usage_km"`
	FuelType           string                `json:"fuel_type"`
	TransmissionType   string                `json:"transmission_type"`
	CurrentStatus      *VehicleStatusSummary `json:"current_status"`
	CreatedAt          string                `json:"created_at"`
	UpdatedAt          string                `json:"updated_at"`
}

type VehicleBuyingSection struct {
	BuyingPrice float64 `json:"buying_price"`
	BuyingDate  string  `json:"buying_date"`
	Currency    string  `json:"currency"`
	Remarks     string  `json:"remarks"`
}

type VehiclePricingSection struct {
	PriceTag float64 `json:"price_tag"`
	TaggedAt string  `json:"tagged_at"`
	Currency string  `json:"currency"`
}

type VehiclePriceTagOnly struct {
	PriceTag float64 `json:"price_tag"`
	Currency string  `json:"currency"`
}

type VehicleStatusSection struct {
	Current *VehicleStatusItem  `json:"current"`
	History []VehicleStatusItem `json:"history"`
}

type VehicleStatusItem struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	StartedAt   string `json:"started_at"`
	EndedAt     string `json:"ended_at"`
}

type VehicleExpenseItem struct {
	ID          uint64  `json:"id"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	PaidTo      string  `json:"paid_to"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type VehicleDocumentItem struct {
	ID           uint64 `json:"id"`
	DocumentType string `json:"document_type"`
	DocumentURL  string `json:"document_url"`
	ValidFrom    string `json:"valid_from"`
	ValidTill    string `json:"valid_till"`
	Remarks      string `json:"remarks"`
	UploadedAt   string `json:"uploaded_at"`
}

type VehicleImageItem struct {
	ID    uint64 `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

type VehicleSaleCustomer struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
}

type VehicleSellingSection struct {
	SalePrice   float64             `json:"sale_price"`
	SaleDate    string              `json:"sale_date"`
	PaymentMode string              `json:"payment_mode"`
	ReceiptUrl  string              `json:"receipt_url"`
	Remarks     string              `json:"remarks"`
	Customer    VehicleSaleCustomer `json:"customer"`
}

type VehicleSoldPriceOnly struct {
	SalePrice float64 `json:"sale_price"`
}

type GetVehicleAdminResponse struct {
	Basic         VehicleBasicSection    `json:"basic"`
	BuyingDetails *VehicleBuyingSection  `json:"buying_details,omitempty"`
	Pricing       *VehiclePricingSection `json:"pricing,omitempty"`
	Status        VehicleStatusSection   `json:"status"`
	Expenses      []VehicleExpenseItem   `json:"expenses"`
	Documents     []VehicleDocumentItem  `json:"documents"`
	Images        []VehicleImageItem     `json:"images"`
	Selling       *VehicleSellingSection `json:"selling,omitempty"`
}

type GetVehicleBasicResponse struct {
	Basic   VehicleBasicSection   `json:"basic"`
	Pricing *VehiclePriceTagOnly  `json:"pricing,omitempty"`
	Selling *VehicleSoldPriceOnly `json:"selling,omitempty"`
}

type CreateVehicleRequest struct {
	VehicleType        VehicleType      `json:"vehicle_type" binding:"required"`
	Manufacturer       string           `json:"manufacturer" binding:"required"`
	Model              string           `json:"model" binding:"required"`
	Variant            string           `json:"variant" binding:"required"`
	Color              string           `json:"color" binding:"required"`
	YearOfManufacture  int              `json:"year_of_manufacture" binding:"required"`
	RTOCode            string           `json:"rto_code" binding:"required"`
	RegistrationNumber string           `json:"registration_number" binding:"required"`
	RegistrationState  string           `json:"registration_state" binding:"required"`
	UsageKM            int              `json:"usage_km" binding:"required"`
	FuelType           FuelType         `json:"fuel_type" binding:"required"`
	TransmissionType   TransmissionType `json:"transmission_type" binding:"required"`
}

type CreateVehicleResponse struct {
	ID                 uint64 `json:"id"`
	VehicleType        string `json:"vehicle_type"`
	Manufacturer       string `json:"manufacturer"`
	Model              string `json:"model"`
	Variant            string `json:"variant"`
	Color              string `json:"color"`
	YearOfManufacture  int    `json:"year_of_manufacture"`
	RTOCode            string `json:"rto_code"`
	RegistrationNumber string `json:"registration_number"`
	RegistrationState  string `json:"registration_state"`
	UsageKM            int    `json:"usage_km"`
	FuelType           string `json:"fuel_type"`
	TransmissionType   string `json:"transmission_type"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}
