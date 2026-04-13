package repositories

import (
	"context"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
)

type PrizeRepository interface {
	Create(ctx context.Context, prize *models.Prize) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Prize, error)
	GetByChallengeID(ctx context.Context, challengeID uuid.UUID) ([]*models.Prize, error)
	Update(ctx context.Context, prize *models.Prize) error
	Delete(ctx context.Context, id uuid.UUID) error
}
