package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware handles structured logging for each request.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		rid, _ := c.Get(RequestIDKey)
		status := c.Writer.Status()
		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("request_id", rid.(string)),
			slog.Duration("duration", duration),
		}

		if status >= http.StatusInternalServerError {
			slog.LogAttrs(c.Request.Context(), slog.LevelError, "request failed", attrs...)
		} else {
			slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "request processed", attrs...)
		}
	}
}
