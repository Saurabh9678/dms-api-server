package user

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateProfileResponse struct {
	Name string `json:"name"`
}
