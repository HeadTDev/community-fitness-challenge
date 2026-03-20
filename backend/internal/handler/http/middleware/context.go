package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDKey = "X-Request-ID"
	UserIDKey    = "user_id"
	RoleKey      = "user_role"
)

// RequestIDMiddleware adds a unique UUID to each request header and context.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDKey)
		if rid == "" {
			rid = uuid.New().String()
		}

		c.Set(RequestIDKey, rid)
		c.Header(RequestIDKey, rid)
		c.Next()
	}
}
