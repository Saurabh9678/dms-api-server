package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"infiour.local/dms-api-server/pkg/logger"
)

func RequestLog(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestLog := base.With(
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		ctx := logger.WithContext(c.Request.Context(), requestLog)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		requestLog.Info("request completed",
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}
