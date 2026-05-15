package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

const (
	headerPlatform = "X-Platform"
	headerDeviceID = "X-Device-Id"
)

var allowedPlatforms = map[string]struct{}{
	"web":            {},
	"ios_mobile":     {},
	"android_mobile": {},
	"desktop":        {},
}

func RequireDeviceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := strings.TrimSpace(c.GetHeader(headerPlatform))
		deviceID := strings.TrimSpace(c.GetHeader(headerDeviceID))

		if platform == "" || deviceID == "" {
			response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidDeviceContext, "invalid request")
			c.Abort()
			return
		}

		if _, ok := allowedPlatforms[platform]; !ok {
			response.Error(c, http.StatusBadRequest, apperrors.CodeInvalidDeviceContext, "invalid request")
			c.Abort()
			return
		}

		c.Next()
	}
}
