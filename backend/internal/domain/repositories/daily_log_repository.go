package repositories

import (
	"context"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
)

type DailyLogRepository interface {
	Create(ctx context.Context, log *models.DailyLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.DailyLog, error)
	GetByUserChallengeDate(ctx context.Context, userID, challengeID uuid.UUID, logDate time.Time) (*models.DailyLog, error)
	ListByUserAndChallenge(ctx context.Context, userID, challengeID uuid.UUID) ([]*models.DailyLog, error)
}

type DailyLogRepositoryWithTx interface {
	DailyLogRepository
	WithTx(tx interface{}) DailyLogRepository
}
