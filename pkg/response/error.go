package response

import (
	"github.com/gin-gonic/gin"
	apperrors "infiour.local/dms-api-server/pkg/errors"
)

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Success bool         `json:"success"`
	Error   errorPayload `json:"error"`
}

func Error(c *gin.Context, status int, code string, msg string) {
	c.JSON(status, ErrorEnvelope{
		Success: false,
		Error: errorPayload{
			Code:    code,
			Message: msg,
		},
	})
}

func FromError(c *gin.Context, err error) {
	appErr := apperrors.ToAppError(err)
	Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
}
