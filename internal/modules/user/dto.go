package user

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateProfileResponse struct {
	Name string `json:"name"`
}

type ShowroomRole struct {
	ShowroomID   uint64       `json:"showroom_id"`
	ShowroomName string       `json:"showroom_name"`
	Role         UserRoleType `json:"role"`
}

type GetProfileResponse struct {
	Name          *string        `json:"name"`
	PhoneNumber   *string        `json:"phone_number"`
	ShowroomRoles []ShowroomRole `json:"showroom_roles"`
}
