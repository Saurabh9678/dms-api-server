package errors

const (
	CodeInternal            = "INTERNAL"
	CodeInvalidRequest      = "INVALID_REQUEST"
	CodeInvalidOTP          = "INVALID_OTP"
	CodeOTPExpired          = "OTP_EXPIRED"
	CodeOTPAlreadyUsed      = "OTP_ALREADY_USED"
	CodeOTPAttemptsExceeded = "OTP_ATTEMPTS_EXCEEDED"
	CodeInvalidRefreshToken = "INVALID_REFRESH_TOKEN"
	CodeSessionRevoked      = "SESSION_REVOKED"
	CodeUserNotFound        = "USER_NOT_FOUND"
)
