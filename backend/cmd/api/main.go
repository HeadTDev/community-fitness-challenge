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
	"github.com/HeadTDev/fitchallenge/internal/handler/http/middleware"
	"github.com/HeadTDev/fitchallenge/internal/pkg/jwt"
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

	// 3. Initialize JWT & AWS Clients
	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, 15*time.Minute, 7*24*time.Hour)
	
	// 4. Initialize Gin router
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(dbPool)
	authHandler := handler.NewAuthHandler(jwtManager)

	// Basic health routes
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register-dev", authHandler.RegisterDev)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected v1 routes
	v1 := r.Group("/v1")
	v1.Use(middleware.AuthMiddleware(jwtManager))
	{
		v1.GET("/users/me", handler.MeHandler)
		v1.GET("/aws-status", healthHandler.AWSStatus)
	}

	// 5. Configure HTTP Server
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
