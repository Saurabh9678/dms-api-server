package auth

type RegisterRequest struct {
	CountryCode string `json:"countryCode" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Platform    string `json:"-"`
	DeviceID    string `json:"-"`
}

type LoginRequest struct {
	CountryCode string `json:"countryCode" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	Platform    string `json:"-"`
	DeviceID    string `json:"-"`
}

type VerifyOTPRequest struct {
	RequestID string `json:"requestId" binding:"required,len=8,alphanum"`
	OTPCode   string `json:"otpCode" binding:"required,len=6,numeric"`
	Platform  string `json:"-"`
	DeviceID  string `json:"-"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	AccessToken string `json:"-"`
	Platform    string `json:"-"`
}

type TriggerOTPResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}
