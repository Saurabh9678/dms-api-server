package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
	"infiour.local/dms-api-server/pkg/response"
)

const ContextKeyUserID = "userID"

type TokenParser interface {
	ParseAccessToken(token string) (uint64, error)
}

func RequireAuth(parser TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := extractBearerToken(c)
		if !ok {
			response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
			c.Abort()
			return
		}

		userID, err := parser.ParseAccessToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, apperrors.CodeInvalidAccessToken, "invalid request")
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, userID)
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) (string, bool) {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return "", false
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if token == "" {
		return "", false
	}

	return token, true
}
