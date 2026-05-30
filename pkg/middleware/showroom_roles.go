package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

const ContextKeyShowroomRoles = "showroomRoles"

// ShowroomRolesProvider loads all showroom roles for a given user.
// Returns a map of showroomID → role string.
type ShowroomRolesProvider interface {
	LoadUserShowroomRoles(ctx context.Context, userID uint64) (map[uint64]string, error)
}

// RequireShowroomRoles loads the authenticated user's showroom roles and stores
// them in the context under ContextKeyShowroomRoles. Handlers use this map to
// check what role the user has in a specific showroom.
func RequireShowroomRoles(provider ShowroomRolesProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get(ContextKeyUserID)
		if !exists {
			response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uint64)
		if !ok {
			response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
			c.Abort()
			return
		}

		roles, err := provider.LoadUserShowroomRoles(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
			c.Abort()
			return
		}

		c.Set(ContextKeyShowroomRoles, roles)
		c.Next()
	}
}
