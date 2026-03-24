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
	"github.com/redis/go-redis/v9"
)

type ChallengeService interface {
	CreateChallenge(ctx context.Context, creatorID uuid.UUID, challenge *models.Challenge) error
	PublishChallenge(ctx context.Context, creatorID uuid.UUID, challengeID uuid.UUID) error
	UploadCoverImage(ctx context.Context, creatorID uuid.UUID, challengeID uuid.UUID, body io.Reader, contentType string) (string, error)
	GetChallenge(ctx context.Context, id uuid.UUID) (*models.Challenge, error)
	ListChallenges(ctx context.Context, status *models.ChallengeStatus) ([]*models.Challenge, error)
	JoinChallenge(ctx context.Context, userID, challengeID uuid.UUID) error
	LeaveChallenge(ctx context.Context, userID, challengeID uuid.UUID) error
}

type challengeService struct {
	challengeRepo     repositories.ChallengeRepository
	userRepo          repositories.UserRepository
	participationRepo repositories.ParticipationRepository
	s3Client          *aws.S3Client
	redisClient       *redis.Client
}

const (
	ChallengeBucket = "fitchallenge-assets"
	ChallengePrefix = "challenges"
)

func NewChallengeService(
	challengeRepo repositories.ChallengeRepository,
	userRepo repositories.UserRepository,
	participationRepo repositories.ParticipationRepository,
	s3Client *aws.S3Client,
	redisClient *redis.Client,
) ChallengeService {
	return &challengeService{
		challengeRepo:     challengeRepo,
		userRepo:          userRepo,
		participationRepo: participationRepo,
		s3Client:          s3Client,
		redisClient:       redisClient,
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

func (s *challengeService) JoinChallenge(ctx context.Context, userID, challengeID uuid.UUID) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	if challenge.Status == models.ChallengeStatusDraft || challenge.Status == models.ChallengeStatusFinished {
		return fmt.Errorf("cannot join a challenge in draft or finished status")
	}

	// Check if already joined
	_, err = s.participationRepo.Get(ctx, userID, challengeID)
	if err == nil {
		return domain.ErrAlreadyExists
	}

	// Check limit
	if challenge.MaxParticipants > 0 && challenge.ParticipantCount >= challenge.MaxParticipants {
		return domain.ErrChallengeFull
	}

	participation := &models.Participation{
		ID:          uuid.New(),
		UserID:      userID,
		ChallengeID: challengeID,
		JoinedAt:    time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.participationRepo.Add(ctx, participation); err != nil {
		return err
	}

	// Increment Redis counter
	counterKey := fmt.Sprintf("challenge_count:%s", challengeID.String())
	if err := s.redisClient.Incr(ctx, counterKey).Err(); err != nil {
		// Log error but don't fail join (will sync later if needed)
		fmt.Printf("Redis error incrementing counter: %v\n", err)
	}

	// Update PostgreSQL counter
	return s.syncParticipantCount(ctx, challengeID)
}

func (s *challengeService) LeaveChallenge(ctx context.Context, userID, challengeID uuid.UUID) error {
	_, err := s.participationRepo.Get(ctx, userID, challengeID)
	if err != nil {
		return err // Likely domain.ErrNotFound
	}

	if err := s.participationRepo.Remove(ctx, userID, challengeID); err != nil {
		return err
	}

	// Decrement Redis counter
	counterKey := fmt.Sprintf("challenge_count:%s", challengeID.String())
	if err := s.redisClient.Decr(ctx, counterKey).Err(); err != nil {
		fmt.Printf("Redis error decrementing counter: %v\n", err)
	}

	// Update PostgreSQL counter
	return s.syncParticipantCount(ctx, challengeID)
}

func (s *challengeService) syncParticipantCount(ctx context.Context, challengeID uuid.UUID) error {
	count, err := s.participationRepo.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		return err
	}

	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	challenge.ParticipantCount = count
	challenge.UpdatedAt = time.Now()

	return s.challengeRepo.Update(ctx, challenge)
}

func (s *challengeService) GetChallenge(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	challenge, err := s.challengeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Try to get count from Redis for speed
	counterKey := fmt.Sprintf("challenge_count:%s", id.String())
	val, err := s.redisClient.Get(ctx, counterKey).Int()
	if err == nil {
		challenge.ParticipantCount = val
	}

	return challenge, nil
}

func (s *challengeService) ListChallenges(ctx context.Context, status *models.ChallengeStatus) ([]*models.Challenge, error) {
	challenges, err := s.challengeRepo.List(ctx, status)
	if err != nil {
		return nil, err
	}

	for _, c := range challenges {
		counterKey := fmt.Sprintf("challenge_count:%s", c.ID.String())
		val, err := s.redisClient.Get(ctx, counterKey).Int()
		if err == nil {
			c.ParticipantCount = val
		}
	}

	return challenges, nil
}
