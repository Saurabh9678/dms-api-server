package vehicle

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
