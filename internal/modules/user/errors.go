package user

import (
	stderrors "errors"
	"net/http"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

var ErrUserNotFound = stderrors.New("user not found")

func init() {
	apperrors.RegisterMapper(func(err error) (*apperrors.AppError, bool) {
		if stderrors.Is(err, ErrUserNotFound) {
			return apperrors.NewAppError(apperrors.CodeUserNotFound, "user not found", http.StatusNotFound, err), true
		}
		return nil, false
	})
}
