package repositories

import (
	"context"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
)

type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, challengeID, userID uuid.UUID, score float64) error
	GetRank(ctx context.Context, challengeID, userID uuid.UUID) (int, error)
	GetTopN(ctx context.Context, challengeID uuid.UUID, limit int64) ([]models.LeaderboardEntry, error)
	GetAroundUser(ctx context.Context, challengeID, userID uuid.UUID, radius int64) ([]models.LeaderboardEntry, error)
	GetTotalCount(ctx context.Context, challengeID uuid.UUID) (int64, error)
}
