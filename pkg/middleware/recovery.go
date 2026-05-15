package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("panic recovered", "panic", recovered)
		response.Error(c, http.StatusInternalServerError, apperrors.CodeInternal, "internal server error")
	})
}
