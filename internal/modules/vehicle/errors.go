package vehicle

import (
	stderrors "errors"
	"net/http"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

var ErrVehicleNotFound = stderrors.New("vehicle not found")
var ErrVehicleSold = stderrors.New("vehicle is sold")

func init() {
	apperrors.RegisterMapper(func(err error) (*apperrors.AppError, bool) {
		if stderrors.Is(err, ErrVehicleNotFound) {
			return apperrors.NewAppError(apperrors.CodeVehicleNotFound, "vehicle not found", http.StatusNotFound, err), true
		}
		if stderrors.Is(err, ErrVehicleSold) {
			return apperrors.NewAppError(apperrors.CodeVehicleUpdateForbidden, "vehicle is sold", http.StatusUnprocessableEntity, err), true
		}
		return nil, false
	})
}
