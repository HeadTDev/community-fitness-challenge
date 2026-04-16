package http

import (
	"context"
	"net/http"

	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	redislib "github.com/redis/go-redis/v9"
)

type redisPinger interface {
	Ping(ctx context.Context) *redislib.StatusCmd
}

type HealthHandler struct {
	db    *pgxpool.Pool
	redis redisPinger
}

func NewHealthHandler(db *pgxpool.Pool, redis redisPinger) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"status":  "ok",
		"service": "api",
	})
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	if err := h.db.Ping(c.Request.Context()); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "DB_NOT_READY", "Database connection is not ready")
		return
	}
	if h.redis == nil {
		response.Error(c, http.StatusServiceUnavailable, "REDIS_NOT_READY", "Redis client is not configured")
		return
	}
	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "REDIS_NOT_READY", "Redis connection is not ready")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"status": "ready",
		"db":     "ok",
		"redis":  "ok",
	})
}

// AWSStatus ellenőrzi az AWS szolgáltatások elérhetőségét (fejlesztéshez).
func (h *HealthHandler) AWSStatus(c *gin.Context) {
	// Egyelőre statikus 'ok' válasz a feladat szerint.
	response.Success(c, http.StatusOK, gin.H{
		"s3":      "ok",
		"sqs":     "ok",
		"secrets": "ok",
	})
}
