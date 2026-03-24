package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	JoinChallenge(ctx context.Context, userID, challengeID uuid.UUID) error
	LeaveChallenge(ctx context.Context, userID, challengeID uuid.UUID) error
}

type challengeService struct {
	dbPool            domain.DBPool
	challengeRepo     repositories.ChallengeRepository
	userRepo          repositories.UserRepository
	participationRepo repositories.ParticipationRepository
	s3Client          aws.S3Client
	redisClient       domain.RedisClient
	logger            *slog.Logger
}

const (
	ChallengeBucket = "fitchallenge-assets"
	ChallengePrefix = "challenges"

	// Redis keys
)

func NewChallengeService(
	dbPool domain.DBPool,
	challengeRepo repositories.ChallengeRepository,
	userRepo repositories.UserRepository,
	participationRepo repositories.ParticipationRepository,
	s3Client aws.S3Client,
	redisClient domain.RedisClient,
	logger *slog.Logger,
) ChallengeService {
	return &challengeService{
		dbPool:            dbPool,
		challengeRepo:     challengeRepo,
		userRepo:          userRepo,
		participationRepo: participationRepo,
		s3Client:          s3Client,
		redisClient:       redisClient,
		logger:            logger,
	}
}

func (s *challengeService) checkCreatorRole(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user for role check", "user_id", userID, "error", err)
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

	// Begin Transaction
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use repositories with transaction
	challengeRepoTx := s.challengeRepo.(repositories.ChallengeRepositoryWithTx).WithTx(tx)
	participationRepoTx := s.participationRepo.(repositories.ParticipationRepositoryWithTx).WithTx(tx)

	if err := participationRepoTx.Add(ctx, participation); err != nil {
		s.logger.Error("failed to add participation", "user_id", userID, "challenge_id", challengeID, "error", err)
		return err
	}

	// Sync count within transaction
	count, err := participationRepoTx.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		return err
	}

	challenge.ParticipantCount = count
	challenge.UpdatedAt = time.Now()
	if err := challengeRepoTx.Update(ctx, challenge); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("commit error: %w", err)
	}

	// Increment Redis counter (after successful DB commit)
	counterKey := fmt.Sprintf(domain.RedisKeyChallengeCount, challengeID.String())
	if err := s.redisClient.Incr(ctx, counterKey).Err(); err != nil {
		s.logger.Warn("Redis error incrementing counter", "challenge_id", challengeID, "error", err)
	}

	return nil
}

func (s *challengeService) LeaveChallenge(ctx context.Context, userID, challengeID uuid.UUID) error {
	_, err := s.participationRepo.Get(ctx, userID, challengeID)
	if err != nil {
		return err // Likely domain.ErrNotFound
	}

	// Begin Transaction
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use repositories with transaction
	challengeRepoTx := s.challengeRepo.(repositories.ChallengeRepositoryWithTx).WithTx(tx)
	participationRepoTx := s.participationRepo.(repositories.ParticipationRepositoryWithTx).WithTx(tx)

	if err := participationRepoTx.Remove(ctx, userID, challengeID); err != nil {
		s.logger.Error("failed to remove participation", "user_id", userID, "challenge_id", challengeID, "error", err)
		return err
	}

	// Sync count within transaction
	count, err := participationRepoTx.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		return err
	}

	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	challenge.ParticipantCount = count
	challenge.UpdatedAt = time.Now()
	if err := challengeRepoTx.Update(ctx, challenge); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("commit error: %w", err)
	}

	// Decrement Redis counter
	counterKey := fmt.Sprintf(domain.RedisKeyChallengeCount, challengeID.String())
	if err := s.redisClient.Decr(ctx, counterKey).Err(); err != nil {
		s.logger.Warn("Redis error decrementing counter", "challenge_id", challengeID, "error", err)
	}

	return nil
}

func (s *challengeService) syncParticipantCount(ctx context.Context, challengeID uuid.UUID) error {
	count, err := s.participationRepo.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		s.logger.Error("failed to get participants count for sync", "challenge_id", challengeID, "error", err)
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
	counterKey := fmt.Sprintf(domain.RedisKeyChallengeCount, id.String())
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
		counterKey := fmt.Sprintf(domain.RedisKeyChallengeCount, c.ID.String())
		val, err := s.redisClient.Get(ctx, counterKey).Int()
		if err == nil {
			c.ParticipantCount = val
		}
	}

	return challenges, nil
}
