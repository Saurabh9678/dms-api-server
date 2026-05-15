package user

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserRoleNotFound     = errors.New("user role not found")
	ErrUserShowroomNotFound = errors.New("user showroom not found")
	ErrInvalidOTP           = errors.New("invalid otp")
	ErrOTPExpired           = errors.New("otp expired")
	ErrOTPAlreadyUsed       = errors.New("otp already used")
	ErrOTPAttemptsExceeded  = errors.New("otp attempts exceeded")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrSessionRevoked       = errors.New("session revoked")
)
