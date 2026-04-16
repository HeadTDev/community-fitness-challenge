package repositories

import (
	"context"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
)

type ChallengeRepository interface {
	Create(ctx context.Context, challenge *models.Challenge) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Challenge, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Challenge, error)
	Update(ctx context.Context, challenge *models.Challenge) error
	UpdateParticipantCount(ctx context.Context, id uuid.UUID, count int, updatedAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, status *models.ChallengeStatus) ([]*models.Challenge, error)
}

type ChallengeRepositoryWithTx interface {
	ChallengeRepository
	WithTx(tx interface{}) ChallengeRepository
}
