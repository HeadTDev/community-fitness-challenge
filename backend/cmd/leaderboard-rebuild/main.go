package main

import (
	"context"
	"fmt"
	"log"

	"github.com/HeadTDev/fitchallenge/internal/adapter/postgres"
	adapterRedis "github.com/HeadTDev/fitchallenge/internal/adapter/redis"
	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	dbPool, err := postgres.NewConnection(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer dbPool.Close()

	redisClient, err := adapterRedis.NewRedisClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	defer redisClient.Close()

	rows, err := dbPool.Query(ctx, `SELECT challenge_id, user_id, current_score FROM participations`)
	if err != nil {
		log.Fatalf("failed to query participations: %v", err)
	}
	defer rows.Close()

	type row struct {
		ChallengeID uuid.UUID
		UserID      uuid.UUID
		Score       int
	}

	byChallenge := make(map[uuid.UUID][]redislib.Z)
	total := 0
	for rows.Next() {
		var r row
		if scanErr := rows.Scan(&r.ChallengeID, &r.UserID, &r.Score); scanErr != nil {
			log.Fatalf("failed to scan participation row: %v", scanErr)
		}
		byChallenge[r.ChallengeID] = append(byChallenge[r.ChallengeID], redislib.Z{
			Score:  float64(r.Score),
			Member: r.UserID.String(),
		})
		total++
	}

	rebuiltKeys := 0
	for challengeID, members := range byChallenge {
		key := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
		if err := redisClient.Del(ctx, key).Err(); err != nil {
			log.Fatalf("failed to clear leaderboard key %s: %v", key, err)
		}
		if len(members) > 0 {
			if err := redisClient.ZAdd(ctx, key, members...).Err(); err != nil {
				log.Fatalf("failed to rebuild leaderboard key %s: %v", key, err)
			}
		}
		rebuiltKeys++
	}

	log.Printf("✅ leaderboard rebuild completed: keys=%d members=%d", rebuiltKeys, total)
}
