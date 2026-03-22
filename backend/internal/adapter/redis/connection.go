package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient initializes a new Redis client based on the provided configuration.
func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     "", // Default is no password in dev environment
		DB:           0,  // use default DB
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Verify connection with a timeout context
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("🔴 Connected to Redis", "host", cfg.RedisHost, "port", cfg.RedisPort)

	return client, nil
}
