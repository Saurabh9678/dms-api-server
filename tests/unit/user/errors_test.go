package user_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/user"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

func TestErrUserNotFoundMapping(t *testing.T) {
	appErr := apperrors.ToAppError(user.ErrUserNotFound)
	assert.NotNil(t, appErr)
	assert.Equal(t, apperrors.CodeUserNotFound, appErr.Code)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPStatus)
}

func TestErrOtherErrorNoMapping(t *testing.T) {
	someErr := errors.New("some other error")
	appErr := apperrors.ToAppError(someErr)
	assert.NotNil(t, appErr)
	assert.NotEqual(t, apperrors.CodeUserNotFound, appErr.Code)
}
