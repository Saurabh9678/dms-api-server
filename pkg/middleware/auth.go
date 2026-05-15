package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidRefreshToken, "unauthorized")
		c.Abort()
	}
}
