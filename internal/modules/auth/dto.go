package auth

type RegisterRequest struct {
	CountryCode string `json:"countryCode" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Platform    string `json:"platform" binding:"required,oneof=web ios_mobile android_mobile desktop"`
	DeviceID    string `json:"deviceId"`
}

type LoginRequest struct {
	CountryCode string `json:"countryCode" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Platform    string `json:"platform" binding:"required,oneof=web ios_mobile android_mobile desktop"`
	DeviceID    string `json:"deviceId"`
}

type VerifyOTPRequest struct {
	CountryCode string `json:"countryCode" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	OTPCode     string `json:"otpCode" binding:"required,len=6,numeric"`
	Platform    string `json:"platform" binding:"required,oneof=web ios_mobile android_mobile desktop"`
	DeviceID    string `json:"deviceId"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type TriggerOTPResponse struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}
