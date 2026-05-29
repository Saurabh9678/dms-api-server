package vehicle

import (
	stderrors "errors"
	"net/http"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

var ErrVehicleNotFound = stderrors.New("vehicle not found")

func init() {
	apperrors.RegisterMapper(func(err error) (*apperrors.AppError, bool) {
		if stderrors.Is(err, ErrVehicleNotFound) {
			return apperrors.NewAppError(apperrors.CodeVehicleNotFound, "vehicle not found", http.StatusNotFound, err), true
		}
		return nil, false
	})
}
