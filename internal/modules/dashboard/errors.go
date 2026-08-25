package dashboard

import (
	"net/http"

	apperrors "infiour.local/dms-api-server/pkg/errors"
)

var ErrInvalidDuration = apperrors.NewAppError(apperrors.CodeInvalidRequest, "invalid request", http.StatusBadRequest, nil)

// ErrForbidden is returned when the caller cannot access the requested showroom dashboard
// (not a member, or role is not owner/manager, or no managed showroom to default to).
var ErrForbidden = apperrors.NewAppError(apperrors.CodeForbidden, "forbidden", http.StatusForbidden, nil)

func isDashboardRole(role string) bool {
	return role == "owner" || role == "manager"
}

// resolveShowroomID picks the dashboard showroom and enforces owner/manager access.
// If showroomID is nil, uses the lowest-ID showroom where the user is owner or manager.
func resolveShowroomID(showroomID *uint64, roles map[uint64]string) (uint64, error) {
	if showroomID != nil {
		role, ok := roles[*showroomID]
		if !ok || !isDashboardRole(role) {
			return 0, ErrForbidden
		}
		return *showroomID, nil
	}

	var (
		best  uint64
		found bool
	)
	for id, role := range roles {
		if !isDashboardRole(role) {
			continue
		}
		if !found || id < best {
			best = id
			found = true
		}
	}
	if !found {
		return 0, ErrForbidden
	}
	return best, nil
}
