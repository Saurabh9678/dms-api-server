package response

import "github.com/gin-gonic/gin"

type SuccessEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, status int, msg string, data any) {
	c.JSON(status, SuccessEnvelope{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func OK(c *gin.Context, msg string, data any) {
	Success(c, 200, msg, data)
}

func Created(c *gin.Context, msg string, data any) {
	Success(c, 201, msg, data)
}
