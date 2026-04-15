package redis

import (
	"context"
	"fmt"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"
)

type LeaderboardRepo struct {
	client *redislib.Client
}

func NewLeaderboardRepo(client *redislib.Client) *LeaderboardRepo {
	return &LeaderboardRepo{client: client}
}

func (r *LeaderboardRepo) UpdateScore(ctx context.Context, challengeID, userID uuid.UUID, score float64) error {
	key := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
	if err := r.client.ZAdd(ctx, key, redislib.Z{
		Score:  score,
		Member: userID.String(),
	}).Err(); err != nil {
		return fmt.Errorf("failed to update leaderboard score: %w", err)
	}
	return nil
}

func (r *LeaderboardRepo) GetRank(ctx context.Context, challengeID, userID uuid.UUID) (int, error) {
	key := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
	rank, err := r.client.ZRevRank(ctx, key, userID.String()).Result()
	if err != nil {
		if err == redislib.Nil {
			return 0, domain.ErrNotFound
		}
		return 0, fmt.Errorf("failed to fetch leaderboard rank: %w", err)
	}
	return int(rank) + 1, nil
}

func (r *LeaderboardRepo) GetTopN(ctx context.Context, challengeID uuid.UUID, limit int64) ([]models.LeaderboardEntry, error) {
	if limit <= 0 {
		return []models.LeaderboardEntry{}, nil
	}

	key := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
	raw, err := r.client.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top leaderboard entries: %w", err)
	}
	return mapRedisEntries(raw, 0)
}

func (r *LeaderboardRepo) GetAroundUser(ctx context.Context, challengeID, userID uuid.UUID, radius int64) ([]models.LeaderboardEntry, error) {
	if radius < 0 {
		radius = 0
	}

	key := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
	rank, err := r.client.ZRevRank(ctx, key, userID.String()).Result()
	if err != nil {
		if err == redislib.Nil {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch user rank for relative leaderboard: %w", err)
	}

	start := int64(rank) - radius
	if start < 0 {
		start = 0
	}
	end := int64(rank) + radius

	raw, err := r.client.ZRevRangeWithScores(ctx, key, start, end).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relative leaderboard entries: %w", err)
	}
	return mapRedisEntries(raw, start)
}

func (r *LeaderboardRepo) GetTotalCount(ctx context.Context, challengeID uuid.UUID) (int64, error) {
	key := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch leaderboard total count: %w", err)
	}
	return count, nil
}

func mapRedisEntries(raw []redislib.Z, start int64) ([]models.LeaderboardEntry, error) {
	entries := make([]models.LeaderboardEntry, 0, len(raw))
	for i, item := range raw {
		member, ok := item.Member.(string)
		if !ok {
			return nil, fmt.Errorf("invalid leaderboard member type %T", item.Member)
		}
		userID, err := uuid.Parse(member)
		if err != nil {
			return nil, fmt.Errorf("invalid leaderboard member uuid: %w", err)
		}
		entries = append(entries, models.LeaderboardEntry{
			UserID: userID,
			Score:  item.Score,
			Rank:   int(start) + i + 1,
		})
	}
	return entries, nil
}
