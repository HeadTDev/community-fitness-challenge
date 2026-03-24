package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log/slog"

	"github.com/HeadTDev/fitchallenge/internal/adapter/postgres"
	"github.com/HeadTDev/fitchallenge/internal/adapter/redis"
	"github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/HeadTDev/fitchallenge/internal/domain/services"
	handler "github.com/HeadTDev/fitchallenge/internal/handler/http"
	"github.com/HeadTDev/fitchallenge/internal/handler/http/middleware"
	"github.com/HeadTDev/fitchallenge/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Initialize Structured Logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 3. Initialize Connections (DB & Redis)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := postgres.NewConnection(ctx, cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	redisClient, err := redis.NewRedisClient(ctx, cfg)
	if err != nil {
		slog.Error("Failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// 4. Initialize JWT & AWS Clients
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, 15*time.Minute, 7*24*time.Hour)
	
	awsCfg, err := aws.NewAWSConfig(ctx, cfg)
	if err != nil {
		slog.Error("Failed to initialize AWS config", "error", err)
		os.Exit(1)
	}
	s3Client := aws.NewS3Client(awsCfg, cfg.S3PublicURL)

	// 5. Initialize Repositories
	userRepo := postgres.NewUserRepo(dbPool)
	challengeRepo := postgres.NewChallengeRepo(dbPool)

	// 6. Initialize Services
	challengeService := services.NewChallengeService(challengeRepo, userRepo, s3Client)

	// 7. Initialize Gin router
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	r := gin.New()
	
	// Use custom middlewares
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.GlobalRateLimit(redisClient))
	r.Use(gin.Recovery())

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(dbPool)
	authHandler := handler.NewAuthHandler(jwtManager, userRepo, cfg.App.Env)
	userHandler := handler.NewUserHandler(userRepo, s3Client)
	challengeHandler := handler.NewChallengeHandler(challengeService)

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
		v1.GET("/users/me", userHandler.MeHandler)
		v1.GET("/users/profile", userHandler.GetProfile)
		v1.PUT("/users/profile", userHandler.UpdateProfile)
		v1.POST("/users/profile/avatar", userHandler.UploadAvatar)
		
		// Challenge routes
		v1.POST("/challenges", challengeHandler.CreateChallenge)
		v1.GET("/challenges", challengeHandler.ListChallenges)
		v1.GET("/challenges/:id", challengeHandler.GetChallenge)
		v1.POST("/challenges/:id/publish", challengeHandler.PublishChallenge)
		v1.POST("/challenges/:id/image", challengeHandler.UploadCoverImage)

		v1.GET("/aws-status", healthHandler.AWSStatus)
	}

	// 5. Configure HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	// 5. Start server in a goroutine
	go func() {
		log.Printf("🚀 API Server starting on port %s", cfg.App.Port)
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

	// 7. Close resources after server is stopped
	dbPool.Close()
	redisClient.Close()

	log.Println("✅ Server stopped cleanly")
}
