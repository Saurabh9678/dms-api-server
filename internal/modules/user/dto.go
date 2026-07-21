package user

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateProfileResponse struct {
	Name string `json:"name"`
}

type ShowroomRole struct {
	ShowroomID         uint64       `json:"showroom_id"`
	ExternalShowroomID string       `json:"external_showroom_id"`
	ShowroomName       string       `json:"showroom_name"`
	Role               UserRoleType `json:"role"`
}

type GetProfileResponse struct {
	Name          *string        `json:"name"`
	CountryCode   *string        `json:"country_code"`
	PhoneNumber   *string        `json:"phone_number"`
	ShowroomRoles []ShowroomRole `json:"showroom_roles"`
	RequiredName  bool           `json:"required_name"`
	HasShowrooms  bool           `json:"has_showrooms"`
	HasVehicles   bool           `json:"has_vehicles"`
}
