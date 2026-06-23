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
