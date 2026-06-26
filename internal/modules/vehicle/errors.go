package vehicle

import (
	stderrors "errors"
	"net/http"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

var ErrVehicleNotFound = stderrors.New("vehicle not found")
var ErrVehicleSold = stderrors.New("vehicle is sold")
var ErrVehicleAlreadyInShowroom = stderrors.New("vehicle already assigned to a showroom")

func init() {
	apperrors.RegisterMapper(func(err error) (*apperrors.AppError, bool) {
		if stderrors.Is(err, ErrVehicleNotFound) {
			return apperrors.NewAppError(apperrors.CodeVehicleNotFound, "vehicle not found", http.StatusNotFound, err), true
		}
		if stderrors.Is(err, ErrVehicleSold) {
			return apperrors.NewAppError(apperrors.CodeVehicleUpdateForbidden, "vehicle is sold", http.StatusUnprocessableEntity, err), true
		}
		if stderrors.Is(err, ErrVehicleAlreadyInShowroom) {
			return apperrors.NewAppError(apperrors.CodeVehicleAlreadyInShowroom, "vehicle already assigned to a showroom", http.StatusConflict, err), true
		}
		return nil, false
	})
}
