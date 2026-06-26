package vehicle_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"infiour.local/dms-api-server/internal/modules/vehicle"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

func TestVehicleErrorMapper(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantCode   string
		wantStatus int
	}{
		{"ErrVehicleNotFound", vehicle.ErrVehicleNotFound, apperrors.CodeVehicleNotFound, http.StatusNotFound},
		{"ErrVehicleSold", vehicle.ErrVehicleSold, apperrors.CodeVehicleUpdateForbidden, http.StatusUnprocessableEntity},
		{"ErrVehicleAlreadyInShowroom", vehicle.ErrVehicleAlreadyInShowroom, apperrors.CodeVehicleAlreadyInShowroom, http.StatusConflict},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			appErr := apperrors.ToAppError(tc.err)
			assert.NotNil(t, appErr)
			assert.Equal(t, tc.wantCode, appErr.Code)
			assert.Equal(t, tc.wantStatus, appErr.HTTPStatus)
		})
	}
}

func TestVehicleErrorMapper_Unknown(t *testing.T) {
	appErr := apperrors.ToAppError(errors.New("some unrelated error"))
	assert.NotNil(t, appErr)
	assert.Equal(t, apperrors.CodeInternal, appErr.Code)
}
