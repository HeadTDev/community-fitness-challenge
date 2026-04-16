package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

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

	challengeRows, err := dbPool.Query(ctx, `SELECT id FROM challenges WHERE deleted_at IS NULL`)
	if err != nil {
		log.Fatalf("failed to query active challenges: %v", err)
	}
	defer challengeRows.Close()

	challengeIDs := make([]uuid.UUID, 0)
	for challengeRows.Next() {
		var challengeID uuid.UUID
		if scanErr := challengeRows.Scan(&challengeID); scanErr != nil {
			log.Fatalf("failed to scan active challenge id: %v", scanErr)
		}
		challengeIDs = append(challengeIDs, challengeID)
	}

	rows, err := dbPool.Query(ctx, `
		SELECT p.challenge_id, p.user_id, p.current_score
		FROM participations p
		INNER JOIN challenges c ON c.id = p.challenge_id
		WHERE c.deleted_at IS NULL
	`)
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
	participantCountByChallenge := make(map[uuid.UUID]int)
	totalParticipants := 0
	for rows.Next() {
		var r row
		if scanErr := rows.Scan(&r.ChallengeID, &r.UserID, &r.Score); scanErr != nil {
			log.Fatalf("failed to scan participation row: %v", scanErr)
		}
		byChallenge[r.ChallengeID] = append(byChallenge[r.ChallengeID], redislib.Z{
			Score:  float64(r.Score),
			Member: r.UserID.String(),
		})
		participantCountByChallenge[r.ChallengeID]++
		totalParticipants++
	}

	counterDriftKeys := 0
	leaderboardDriftKeys := 0
	for _, challengeID := range challengeIDs {
		desiredCount := participantCountByChallenge[challengeID]
		counterKey := fmt.Sprintf(domain.RedisKeyChallengeCount, challengeID.String())
		currentCount, getCountErr := redisClient.Get(ctx, counterKey).Int()
		if getCountErr != nil && getCountErr != redislib.Nil {
			log.Fatalf("failed to read counter key %s: %v", counterKey, getCountErr)
		}
		if getCountErr == nil && currentCount != desiredCount {
			counterDriftKeys++
		}
		if err := redisClient.Set(ctx, counterKey, desiredCount, 0).Err(); err != nil {
			log.Fatalf("failed to sync counter key %s: %v", counterKey, err)
		}

		leaderboardKey := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
		members := byChallenge[challengeID]
		sort.SliceStable(members, func(i, j int) bool {
			if members[i].Score == members[j].Score {
				return members[i].Member.(string) < members[j].Member.(string)
			}
			return members[i].Score > members[j].Score
		})

		cardinality, cardErr := redisClient.ZCard(ctx, leaderboardKey).Result()
		if cardErr != nil {
			log.Fatalf("failed to read leaderboard cardinality for %s: %v", leaderboardKey, cardErr)
		}
		if cardinality != int64(len(members)) {
			leaderboardDriftKeys++
		}

		if len(members) == 0 {
			if err := redisClient.Del(ctx, leaderboardKey).Err(); err != nil {
				log.Fatalf("failed to clear empty leaderboard key %s: %v", leaderboardKey, err)
			}
			continue
		}

		stagingKey := fmt.Sprintf("%s:rebuild:%d", leaderboardKey, time.Now().UnixNano())
		if err := redisClient.Del(ctx, stagingKey).Err(); err != nil {
			log.Fatalf("failed to clear staging leaderboard key %s: %v", stagingKey, err)
		}
		if err := redisClient.ZAdd(ctx, stagingKey, members...).Err(); err != nil {
			log.Fatalf("failed to rebuild staging leaderboard key %s: %v", stagingKey, err)
		}
		if err := redisClient.Rename(ctx, stagingKey, leaderboardKey).Err(); err != nil {
			log.Fatalf("failed to swap staging leaderboard key %s into %s: %v", stagingKey, leaderboardKey, err)
		}
	}

	log.Printf(
		"✅ leaderboard and counter rebuild completed: challenges=%d participants=%d counter_drift_keys=%d leaderboard_drift_keys=%d",
		len(challengeIDs),
		totalParticipants,
		counterDriftKeys,
		leaderboardDriftKeys,
	)
}
