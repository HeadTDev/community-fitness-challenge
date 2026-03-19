package http

import (
	"net/http"

	"github.com/HeadTDev/fitchallenge/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"status":  "ok",
		"service": "api",
	})
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	// TODO: Add real DB and Redis connection checks
	response.Success(c, http.StatusOK, gin.H{
		"status":  "ready",
		"db":      "ok",
		"redis":   "ok",
	})
}
