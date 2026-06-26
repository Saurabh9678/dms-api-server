package showroom

import "encoding/json"

type CreateShowroomRequest struct {
	Name        string `form:"name"`
	Geolocation string `form:"geolocation"`
}

type CreateShowroomResponse struct {
	ID             uint64          `json:"id"`
	Name           string          `json:"name"`
	ShowroomLogo   *string         `json:"showroom_logo"`
	ShowroomBanner *string         `json:"showroom_banner"`
	Geolocation    json.RawMessage `json:"geolocation,omitempty"`
}

type AddMemberRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

type AddMemberResponse struct {
	ShowroomID uint64 `json:"showroom_id"`
	UserID     uint64 `json:"user_id"`
	Role       string `json:"role"`
}

type MemberItem struct {
	UserID      uint64  `json:"user_id"`
	Name        *string `json:"name"`
	PhoneNumber *string `json:"phone_number"`
	Role        string  `json:"role"`
}

type ListMembersResponse struct {
	Members []MemberItem `json:"members"`
	Total   int64        `json:"total"`
	Page    int          `json:"page"`
	Limit   int          `json:"limit"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type UpdateShowroomRequest struct {
	Name         string `form:"name"`
	Geolocation  string `form:"geolocation"`
	RemoveLogo   string `form:"remove_logo"`
	RemoveBanner string `form:"remove_banner"`
}
