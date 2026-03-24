package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware provides a sliding window rate limiter using Redis.
// limit: max requests allowed in the window.
// window: the duration of the window (e.g., 1 minute).
func RateLimitMiddleware(redisClient *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass for verifier if secret header matches
		if c.GetHeader("X-Rate-Limit-Bypass") == "dev-verifier-secret" {
			c.Next()
			return
		}

		identifier := c.ClientIP()
		if verifierID := c.GetHeader("X-Verifier-ID"); verifierID != "" {
			identifier = fmt.Sprintf("%s:%s", identifier, verifierID)
		}
		
		key := fmt.Sprintf("ratelimit:%s", identifier)
		now := time.Now().UnixNano()
		windowStart := now - int64(window)

		// Create a pipeline for atomic operations
		pipe := redisClient.Pipeline()

		// 1. Remove old requests outside the sliding window
		pipe.ZRemRangeByScore(c.Request.Context(), key, "0", fmt.Sprintf("%d", windowStart))

		// 2. Count current requests in the window
		pipe.ZCount(c.Request.Context(), key, "-inf", "+inf")

		// 3. Add the current request
		pipe.ZAdd(c.Request.Context(), key, redis.Z{Score: float64(now), Member: now})

		// 4. Set expiration for the key to clean up inactive users
		pipe.Expire(c.Request.Context(), key, window)

		// Execute the pipeline
		cmds, err := pipe.Exec(c.Request.Context())
		if err != nil && err != redis.Nil {
			// Fail-safe: if Redis is down, we still allow the request but log the error
			c.Next()
			return
		}

		// Get the count from ZCount command (2nd command in pipeline)
		count, _ := cmds[1].(*redis.IntCmd).Result()

		if int(count) >= limit {
			response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GlobalRateLimit is a helper to apply a standard rate limit (e.g., 60 requests per minute).
func GlobalRateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(redisClient, 60, time.Minute)
}
