package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/auth"
)

func TestUserOTPTableName(t *testing.T) {
	assert.Equal(t, "user_otps", auth.UserOTP{}.TableName())
}

func TestUserSessionTableName(t *testing.T) {
	assert.Equal(t, "user_sessions", auth.UserSession{}.TableName())
}
