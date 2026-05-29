package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		c.Header(requestIDHeader, id)
		c.Set("request_id", id)
		c.Next()
	}
}

func newRequestID() string {
	buf := make([]byte, 16)
	// crypto/rand.Read panics on entropy failure in Go 1.20+; error is unreachable.
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
