package http

import (
	"net/http"

	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"status":  "ok",
		"service": "api",
	})
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	// Senior tip: Mindig ellenőrizzük az adatbázis kapcsolatot a readyz végponton.
	if err := h.db.Ping(c.Request.Context()); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "DB_NOT_READY", "Database connection is not ready")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"status": "ready",
		"db":     "ok",
		"redis":  "ok", // Redis kliens majd később kerül beépítésre
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
