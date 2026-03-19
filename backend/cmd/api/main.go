package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/adapter/postgres"
	"github.com/HeadTDev/fitchallenge/internal/config"
	handler "github.com/HeadTDev/fitchallenge/internal/handler/http"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Initialize Database Connection
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := postgres.NewConnection(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("🐘 Closing database connection pool...")
		dbPool.Close()
	}()

	// 3. Initialize Gin router
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(dbPool)

	// Basic routes
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	// 4. Configure HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	// 5. Start server in a goroutine
	go func() {
		log.Printf("🚀 API Server starting on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// 6. Wait for interrupt signal for graceful shutdown
	<-ctx.Done()

	log.Println("⚠️  Shutting down server gracefully...")

	// Create a timeout context for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
