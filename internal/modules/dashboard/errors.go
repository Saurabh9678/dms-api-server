package dashboard

import (
	"net/http"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

var ErrInvalidDuration = apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)
