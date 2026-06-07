package auth

import (
	stderrors "errors"
	"net/http"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

var (
	ErrInvalidAccessToken   = stderrors.New("invalid access token")
	ErrInvalidOTP           = stderrors.New("invalid otp")
	ErrOTPExpired           = stderrors.New("otp expired")
	ErrOTPAlreadyUsed       = stderrors.New("otp already used")
	ErrOTPAttemptsExceeded  = stderrors.New("otp attempts exceeded")
	ErrInvalidRefreshToken  = stderrors.New("invalid refresh token")
	ErrSessionRevoked       = stderrors.New("session revoked")
	ErrOTPCooldown          = stderrors.New("otp cooldown active")
	ErrOTPRateLimitExceeded = stderrors.New("otp rate limit exceeded")
)

func init() {
	apperrors.RegisterMapper(func(err error) (*apperrors.AppError, bool) {
		switch {
		case stderrors.Is(err, ErrInvalidAccessToken):
			return apperrors.NewAppError(apperrors.CodeInvalidAccessToken, "Invalid access token", http.StatusUnauthorized, err), true
		case stderrors.Is(err, ErrInvalidOTP):
			return apperrors.NewAppError(apperrors.CodeInvalidOTP, "Invalid OTP", http.StatusUnauthorized, err), true
		case stderrors.Is(err, ErrOTPExpired):
			return apperrors.NewAppError(apperrors.CodeOTPExpired, "OTP expired", http.StatusUnauthorized, err), true
		case stderrors.Is(err, ErrOTPAlreadyUsed):
			return apperrors.NewAppError(apperrors.CodeOTPAlreadyUsed, "OTP already used", http.StatusUnauthorized, err), true
		case stderrors.Is(err, ErrOTPAttemptsExceeded):
			return apperrors.NewAppError(apperrors.CodeOTPAttemptsExceeded, "OTP attempts exceeded", http.StatusUnauthorized, err), true
		case stderrors.Is(err, ErrInvalidRefreshToken):
			return apperrors.NewAppError(apperrors.CodeInvalidRefreshToken, "Invalid refresh token", http.StatusUnauthorized, err), true
		case stderrors.Is(err, ErrSessionRevoked):
			return apperrors.NewAppError(apperrors.CodeSessionRevoked, "Session revoked", http.StatusUnauthorized, err), true
		case stderrors.Is(err, ErrOTPCooldown):
			return apperrors.NewAppError(apperrors.CodeOTPCooldown, "please wait before requesting another OTP", http.StatusTooManyRequests, err), true
		case stderrors.Is(err, ErrOTPRateLimitExceeded):
			return apperrors.NewAppError(apperrors.CodeOTPRateLimitExceeded, "too many OTP requests, please try again later", http.StatusTooManyRequests, err), true
		default:
			return nil, false
		}
	})
}
