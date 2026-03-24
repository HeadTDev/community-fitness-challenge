package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/aws"
	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/google/uuid"
)

type ChallengeService interface {
	CreateChallenge(ctx context.Context, creatorID uuid.UUID, challenge *models.Challenge) error
	PublishChallenge(ctx context.Context, creatorID uuid.UUID, challengeID uuid.UUID) error
	UploadCoverImage(ctx context.Context, creatorID uuid.UUID, challengeID uuid.UUID, body io.Reader, contentType string) (string, error)
	GetChallenge(ctx context.Context, id uuid.UUID) (*models.Challenge, error)
	ListChallenges(ctx context.Context, status *models.ChallengeStatus) ([]*models.Challenge, error)
}

type challengeService struct {
	challengeRepo repositories.ChallengeRepository
	userRepo      repositories.UserRepository
	s3Client      *aws.S3Client
}

const (
	ChallengeBucket = "fitchallenge-assets"
	ChallengePrefix = "challenges"
)

func NewChallengeService(
	challengeRepo repositories.ChallengeRepository,
	userRepo repositories.UserRepository,
	s3Client *aws.S3Client,
) ChallengeService {
	return &challengeService{
		challengeRepo: challengeRepo,
		userRepo:      userRepo,
		s3Client:      s3Client,
	}
}

func (s *challengeService) checkCreatorRole(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Role != models.RoleCreator && user.Role != models.RoleAdmin {
		return domain.ErrUnauthorized // Or a more specific error like ErrForbidden
	}

	return nil
}

func (s *challengeService) CreateChallenge(ctx context.Context, creatorID uuid.UUID, challenge *models.Challenge) error {
	if err := s.checkCreatorRole(ctx, creatorID); err != nil {
		return err
	}

	challenge.ID = uuid.New()
	challenge.CreatorID = creatorID
	challenge.Status = models.ChallengeStatusDraft
	challenge.CreatedAt = time.Now()
	challenge.UpdatedAt = time.Now()

	return s.challengeRepo.Create(ctx, challenge)
}

func (s *challengeService) PublishChallenge(ctx context.Context, creatorID uuid.UUID, challengeID uuid.UUID) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	// Only the creator or an admin can publish
	if challenge.CreatorID != creatorID {
		user, err := s.userRepo.GetByID(ctx, creatorID)
		if err != nil || user.Role != models.RoleAdmin {
			return domain.ErrUnauthorized
		}
	}

	if challenge.Status != models.ChallengeStatusDraft {
		return fmt.Errorf("only draft challenges can be published")
	}

	challenge.Status = models.ChallengeStatusUpcoming
	challenge.UpdatedAt = time.Now()

	return s.challengeRepo.Update(ctx, challenge)
}

func (s *challengeService) UploadCoverImage(ctx context.Context, creatorID uuid.UUID, challengeID uuid.UUID, body io.Reader, contentType string) (string, error) {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return "", err
	}

	// Authorization check
	if challenge.CreatorID != creatorID {
		user, err := s.userRepo.GetByID(ctx, creatorID)
		if err != nil || user.Role != models.RoleAdmin {
			return "", domain.ErrUnauthorized
		}
	}

	key := fmt.Sprintf("%s/%s/cover_%d", ChallengePrefix, challengeID.String(), time.Now().Unix())
	url, err := s.s3Client.UploadFile(ctx, ChallengeBucket, key, body, contentType)
	if err != nil {
		return "", err
	}

	challenge.ImageURL = &url
	challenge.UpdatedAt = time.Now()

	if err := s.challengeRepo.Update(ctx, challenge); err != nil {
		return "", err
	}

	return url, nil
}

func (s *challengeService) GetChallenge(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	return s.challengeRepo.GetByID(ctx, id)
}

func (s *challengeService) ListChallenges(ctx context.Context, status *models.ChallengeStatus) ([]*models.Challenge, error) {
	return s.challengeRepo.List(ctx, status)
}
