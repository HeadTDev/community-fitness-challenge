package repositories

import (
	"context"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
)

type ParticipationRepository interface {
	Add(ctx context.Context, p *models.Participation) error
	Remove(ctx context.Context, userID, challengeID uuid.UUID) error
	Get(ctx context.Context, userID, challengeID uuid.UUID) (*models.Participation, error)
	GetParticipantsCount(ctx context.Context, challengeID uuid.UUID) (int, error)
	ListByChallenge(ctx context.Context, challengeID uuid.UUID) ([]*models.Participation, error)
}
