package auth_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/auth"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

func TestAuthErrorMapper_AllCases(t *testing.T) {
	cases := []struct {
		name           string
		err            error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "ErrInvalidAccessToken",
			err:            auth.ErrInvalidAccessToken,
			expectedCode:   apperrors.CodeInvalidAccessToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrInvalidOTP",
			err:            auth.ErrInvalidOTP,
			expectedCode:   apperrors.CodeInvalidOTP,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrOTPExpired",
			err:            auth.ErrOTPExpired,
			expectedCode:   apperrors.CodeOTPExpired,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrOTPAlreadyUsed",
			err:            auth.ErrOTPAlreadyUsed,
			expectedCode:   apperrors.CodeOTPAlreadyUsed,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrOTPAttemptsExceeded",
			err:            auth.ErrOTPAttemptsExceeded,
			expectedCode:   apperrors.CodeOTPAttemptsExceeded,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrInvalidRefreshToken",
			err:            auth.ErrInvalidRefreshToken,
			expectedCode:   apperrors.CodeInvalidRefreshToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrSessionRevoked",
			err:            auth.ErrSessionRevoked,
			expectedCode:   apperrors.CodeSessionRevoked,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrOTPCooldown",
			err:            auth.ErrOTPCooldown,
			expectedCode:   apperrors.CodeOTPCooldown,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "ErrOTPRateLimitExceeded",
			err:            auth.ErrOTPRateLimitExceeded,
			expectedCode:   apperrors.CodeOTPRateLimitExceeded,
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			appErr := apperrors.ToAppError(tc.err)
			assert.NotNil(t, appErr)
			assert.Equal(t, tc.expectedCode, appErr.Code)
			assert.Equal(t, tc.expectedStatus, appErr.HTTPStatus)
		})
	}
}
