package main

import (
	"log"

	"github.com/HeadTDev/fitchallenge/internal/config"
	handler "github.com/HeadTDev/fitchallenge/internal/handler/http"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Gin router
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()

	// Basic routes
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	// Start server
	log.Printf("🚀 API Server starting on port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
