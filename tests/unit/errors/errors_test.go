package errors_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

func TestAppError_Error_WithWrapped(t *testing.T) {
	wrapped := errors.New("underlying cause")
	e := apperrors.NewAppError("CODE", "message", http.StatusBadRequest, wrapped)
	assert.Equal(t, "message: underlying cause", e.Error())
}

func TestAppError_Error_WithoutWrapped(t *testing.T) {
	e := apperrors.NewAppError("CODE", "message", http.StatusBadRequest, nil)
	assert.Equal(t, "message", e.Error())
}

func TestAppError_Error_Nil(t *testing.T) {
	var e *apperrors.AppError
	assert.Equal(t, "", e.Error())
}

func TestAppError_Unwrap_NonNil(t *testing.T) {
	wrapped := errors.New("cause")
	e := apperrors.NewAppError("CODE", "msg", http.StatusBadRequest, wrapped)
	assert.Equal(t, wrapped, e.Unwrap())
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	var e *apperrors.AppError
	assert.Nil(t, e.Unwrap())
}

func TestAppError_Unwrap_NoWrapped(t *testing.T) {
	e := apperrors.NewAppError("CODE", "msg", http.StatusBadRequest, nil)
	assert.Nil(t, e.Unwrap())
}

func TestToAppError_NilError(t *testing.T) {
	result := apperrors.ToAppError(nil)
	assert.Nil(t, result)
}

func TestToAppError_AppError(t *testing.T) {
	original := apperrors.NewAppError("MY_CODE", "my message", http.StatusConflict, nil)
	result := apperrors.ToAppError(original)
	assert.Equal(t, original, result)
}

func TestToAppError_UnmappedError(t *testing.T) {
	err := errors.New("random error")
	result := apperrors.ToAppError(err)
	assert.NotNil(t, result)
	assert.Equal(t, apperrors.CodeInternal, result.Code)
	assert.Equal(t, http.StatusInternalServerError, result.HTTPStatus)
}

func TestToAppError_MappedError(t *testing.T) {
	sentinel := errors.New("sentinel")
	apperrors.RegisterMapper(func(err error) (*apperrors.AppError, bool) {
		if errors.Is(err, sentinel) {
			return apperrors.NewAppError("SENTINEL", "sentinel error", http.StatusTeapot, err), true
		}
		return nil, false
	})

	result := apperrors.ToAppError(sentinel)
	assert.NotNil(t, result)
	assert.Equal(t, "SENTINEL", result.Code)
	assert.Equal(t, http.StatusTeapot, result.HTTPStatus)
}
