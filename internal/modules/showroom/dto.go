package showroom

import "encoding/json"

type CreateShowroomRequest struct {
	Name        string `form:"name"`
	Geolocation string `form:"geolocation"`
}

type CreateShowroomResponse struct {
	ID             uint64          `json:"id"`
	ShowroomID     string          `json:"showroom_id"`
	Name           string          `json:"name"`
	ShowroomLogo   *string         `json:"showroom_logo"`
	ShowroomBanner *string         `json:"showroom_banner"`
	Geolocation    json.RawMessage `json:"geolocation,omitempty"`
}

type GetShowroomResponse struct {
	ID             uint64          `json:"id"`
	ShowroomID     string          `json:"showroom_id"`
	Name           string          `json:"name"`
	ShowroomLogo   *string         `json:"showroom_logo"`
	ShowroomBanner *string         `json:"showroom_banner"`
	Geolocation    json.RawMessage `json:"geolocation,omitempty"`
	Role           string          `json:"role"`
}

type ListShowroomsResponse struct {
	Showrooms []ShowroomListItem `json:"showrooms"`
}

type ShowroomListItem struct {
	ID         uint64 `json:"id"`
	ShowroomID string `json:"showroom_id"`
	Name       string `json:"name"`
	Role       string `json:"role"`
}

type AddMemberRequest struct {
	Name        string `json:"name" binding:"required"`
	CountryCode string `json:"country_code" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Role        string `json:"role" binding:"required"`
}

type AddMemberResponse struct {
	ShowroomID uint64 `json:"showroom_id"`
	UserID     uint64 `json:"user_id"`
	Role       string `json:"role"`
}

type MemberItem struct {
	UserID      uint64  `json:"user_id"`
	Name        *string `json:"name"`
	CountryCode *string `json:"country_code"`
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
