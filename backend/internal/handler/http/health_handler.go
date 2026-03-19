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
	dbStatus := "ok"
	if err := h.db.Ping(c.Request.Context()); err != nil {
		dbStatus = "error"
	}

	response.Success(c, http.StatusOK, gin.H{
		"status": "ready",
		"db":     dbStatus,
		"redis":  "ok", // Redis kliens majd később kerül beépítésre
	})
}
