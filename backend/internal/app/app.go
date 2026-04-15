package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/adapter/postgres"
	adapterRedis "github.com/HeadTDev/fitchallenge/internal/adapter/redis"
	"github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/services"
	handler "github.com/HeadTDev/fitchallenge/internal/handler/http"
	"github.com/HeadTDev/fitchallenge/internal/handler/http/middleware"
	"github.com/HeadTDev/fitchallenge/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Container holds all initialized dependencies
type Container struct {
	Config      *config.Config
	Logger      *slog.Logger
	DBPool      *pgxpool.Pool
	RedisClient domain.RedisClient
	JWTManager  *jwt.JWTManager
	S3Client    aws.S3Client
}

// NewRouter initializes the whole dependency tree and returns the gin router
func NewRouter(cfg *config.Config, logger *slog.Logger, ctx context.Context) (*gin.Engine, func(), error) {
	// 1. Initialize Connections
	dbPool, err := postgres.NewConnection(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	redisClient, err := adapterRedis.NewRedisClient(ctx, cfg)
	if err != nil {
		dbPool.Close()
		return nil, nil, err
	}
	redisNative := redisClient

	cleanup := func() {
		dbPool.Close()
		redisClient.Close()
	}

	// 2. Initialize JWT & AWS Clients
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, 15*time.Minute, 7*24*time.Hour)

	awsCfg, err := aws.NewAWSConfig(ctx, cfg)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	s3Client := aws.NewS3Client(awsCfg, cfg.S3PublicURL)
	sqsClient := aws.NewSQSClient(awsCfg)

	// 3. Initialize Repositories
	userRepo := postgres.NewUserRepo(dbPool)
	challengeRepo := postgres.NewChallengeRepo(dbPool)
	participationRepo := postgres.NewParticipationRepo(dbPool)
	prizeRepo := postgres.NewPrizeRepo(dbPool)
	dailyLogRepo := postgres.NewDailyLogRepo(dbPool)
	leaderboardRepo := adapterRedis.NewLeaderboardRepo(redisNative)
	leaderboardFallbackRepo := postgres.NewLeaderboardRepo(dbPool)

	// 4. Initialize Services
	challengeService := services.NewChallengeService(dbPool, challengeRepo, userRepo, participationRepo, prizeRepo, s3Client, redisClient, logger)
	scoringService := services.NewScoringService()
	logService := services.NewLogService(challengeRepo, participationRepo, dailyLogRepo, redisClient, scoringService, sqsClient, logger)
	leaderboardService := services.NewLeaderboardService(leaderboardRepo, leaderboardFallbackRepo, participationRepo, challengeRepo)

	// 5. Initialize Router
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.GlobalRateLimit(redisClient))
	r.Use(gin.Recovery())

	// Handlers
	healthHandler := handler.NewHealthHandler(dbPool)
	authHandler := handler.NewAuthHandler(jwtManager, userRepo, cfg.App.Env)
	userHandler := handler.NewUserHandler(userRepo, s3Client)
	challengeHandler := handler.NewChallengeHandler(challengeService, logService, leaderboardService)

	// Register Routes
	RegisterRoutes(r, healthHandler, authHandler, userHandler, challengeHandler, jwtManager)

	return r, cleanup, nil
}

func RegisterRoutes(
	r *gin.Engine,
	health *handler.HealthHandler,
	auth *handler.AuthHandler,
	user *handler.UserHandler,
	challenge *handler.ChallengeHandler,
	jwtManager *jwt.JWTManager,
) {
	// Basic health routes
	r.GET("/healthz", health.Healthz)
	r.GET("/readyz", health.Readyz)

	// Auth routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register-dev", auth.RegisterDev)
		authGroup.POST("/refresh", auth.RefreshToken)
	}

	// Protected v1 routes
	v1 := r.Group("/v1")
	v1.Use(middleware.AuthMiddleware(jwtManager))
	{
		v1.GET("/users/me", user.MeHandler)
		v1.GET("/users/profile", user.GetProfile)
		v1.PUT("/users/profile", user.UpdateProfile)
		v1.POST("/users/profile/avatar", user.UploadAvatar)

		// Challenge routes
		v1.POST("/challenges", challenge.CreateChallenge)
		v1.GET("/challenges", challenge.ListChallenges)
		v1.GET("/challenges/:id", challenge.GetChallenge)
		v1.POST("/challenges/:id/publish", challenge.PublishChallenge)
		v1.POST("/challenges/:id/image", challenge.UploadCoverImage)
		v1.POST("/challenges/:id/join", challenge.JoinChallenge)
		v1.POST("/challenges/:id/leave", challenge.LeaveChallenge)
		v1.POST("/challenges/:id/logs", challenge.SubmitDailyLog)
		v1.GET("/challenges/:id/logs", challenge.GetDailyLogs)
		v1.GET("/challenges/:id/my-progress", challenge.GetMyProgress)
		v1.GET("/challenges/:id/leaderboard", challenge.GetLeaderboard)
		v1.GET("/challenges/:id/leaderboard/relative", challenge.GetRelativeLeaderboard)

		// Prize routes
		v1.GET("/challenges/:id/prizes", challenge.GetPrizes)
		v1.POST("/challenges/:id/prizes", challenge.AddPrize)
		v1.PUT("/challenges/:id/prizes/:prize_id", challenge.UpdatePrize)
		v1.DELETE("/challenges/:id/prizes/:prize_id", challenge.DeletePrize)

		v1.GET("/aws-status", health.AWSStatus)
	}
}
