package repositories

import (
	"context"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
)

type ChallengeRepository interface {
	Create(ctx context.Context, challenge *models.Challenge) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Challenge, error)
	Update(ctx context.Context, challenge *models.Challenge) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, status *models.ChallengeStatus) ([]*models.Challenge, error)
}

type ChallengeRepositoryWithTx interface {
	ChallengeRepository
	WithTx(tx interface{}) ChallengeRepository
}
