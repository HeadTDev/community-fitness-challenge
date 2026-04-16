package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
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

	// Prize methods
	AddPrize(ctx context.Context, creatorID, challengeID uuid.UUID, prize *models.Prize) error
	UpdatePrize(ctx context.Context, creatorID, challengeID, prizeID uuid.UUID, prize *models.Prize) error
	DeletePrize(ctx context.Context, creatorID, challengeID, prizeID uuid.UUID) error
	GetPrizesByChallengeID(ctx context.Context, challengeID uuid.UUID) ([]*models.Prize, error)
}

type challengeService struct {
	dbPool            domain.DBPool
	challengeRepo     repositories.ChallengeRepository
	userRepo          repositories.UserRepository
	participationRepo repositories.ParticipationRepository
	prizeRepo         repositories.PrizeRepository
	s3Client          aws.S3Client
	redisClient       domain.RedisClient
	sqsClient         challengeEventPublisher
	logger            *slog.Logger
}

const (
	ChallengeBucket          = "fitchallenge-assets"
	ChallengePrefix          = "challenges"
	challengeEventsQueueName = "fitchallenge-jobs"
	notificationSenderEmail  = "noreply@fitchallenge.local"

	// Redis keys
)

type challengeEventPublisher interface {
	GetQueueURL(ctx context.Context, queueName string) (string, error)
	SendMessage(ctx context.Context, queueURL, body string) (string, error)
}

func NewChallengeService(
	dbPool domain.DBPool,
	challengeRepo repositories.ChallengeRepository,
	userRepo repositories.UserRepository,
	participationRepo repositories.ParticipationRepository,
	prizeRepo repositories.PrizeRepository,
	s3Client aws.S3Client,
	redisClient domain.RedisClient,
	sqsClient challengeEventPublisher,
	logger *slog.Logger,
) ChallengeService {
	return &challengeService{
		dbPool:            dbPool,
		challengeRepo:     challengeRepo,
		userRepo:          userRepo,
		participationRepo: participationRepo,
		prizeRepo:         prizeRepo,
		s3Client:          s3Client,
		redisClient:       redisClient,
		sqsClient:         sqsClient,
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

	if err := s.challengeRepo.Update(ctx, challenge); err != nil {
		return err
	}

	s.enqueuePublishNotifications(ctx, challenge)
	return nil
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
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use repositories with transaction
	challengeRepoTx := s.challengeRepo.(repositories.ChallengeRepositoryWithTx).WithTx(tx)
	participationRepoTx := s.participationRepo.(repositories.ParticipationRepositoryWithTx).WithTx(tx)

	challenge, err := challengeRepoTx.GetByIDForUpdate(ctx, challengeID)
	if err != nil {
		return err
	}

	if challenge.Status == models.ChallengeStatusDraft || challenge.Status == models.ChallengeStatusFinished {
		return fmt.Errorf("cannot join a challenge in draft or finished status")
	}

	// Check if already joined
	_, err = participationRepoTx.Get(ctx, userID, challengeID)
	if err == nil {
		return domain.ErrAlreadyExists
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	// Check capacity under row lock
	currentCount, err := participationRepoTx.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		return err
	}
	if challenge.MaxParticipants > 0 && currentCount >= challenge.MaxParticipants {
		return domain.ErrChallengeFull
	}

	participation := &models.Participation{
		ID:          uuid.New(),
		UserID:      userID,
		ChallengeID: challengeID,
		JoinedAt:    time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := participationRepoTx.Add(ctx, participation); err != nil {
		s.logger.Error("failed to add participation", "user_id", userID, "challenge_id", challengeID, "error", err)
		return err
	}

	// Sync count within transaction
	count, err := participationRepoTx.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		return err
	}

	if err := challengeRepoTx.UpdateParticipantCount(ctx, challengeID, count, time.Now()); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("commit error: %w", err)
	}

	s.syncChallengeCountCache(ctx, challengeID, count)

	return nil
}

func (s *challengeService) LeaveChallenge(ctx context.Context, userID, challengeID uuid.UUID) error {
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

	if _, err := participationRepoTx.Get(ctx, userID, challengeID); err != nil {
		return err
	}
	if _, err := challengeRepoTx.GetByIDForUpdate(ctx, challengeID); err != nil {
		return err
	}

	if err := participationRepoTx.Remove(ctx, userID, challengeID); err != nil {
		s.logger.Error("failed to remove participation", "user_id", userID, "challenge_id", challengeID, "error", err)
		return err
	}

	// Sync count within transaction
	count, err := participationRepoTx.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		return err
	}

	if err := challengeRepoTx.UpdateParticipantCount(ctx, challengeID, count, time.Now()); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("commit error: %w", err)
	}

	s.syncChallengeCountCache(ctx, challengeID, count)

	leaderboardKey := fmt.Sprintf(domain.RedisKeyLeaderboard, challengeID.String())
	if err := s.redisClient.Del(ctx, leaderboardKey).Err(); err != nil {
		s.logger.Warn("Redis error clearing leaderboard cache on leave", "challenge_id", challengeID, "error", err)
	}

	return nil
}

func (s *challengeService) syncParticipantCount(ctx context.Context, challengeID uuid.UUID) error {
	count, err := s.participationRepo.GetParticipantsCount(ctx, challengeID)
	if err != nil {
		s.logger.Error("failed to get participants count for sync", "challenge_id", challengeID, "error", err)
		return err
	}

	if err := s.challengeRepo.UpdateParticipantCount(ctx, challengeID, count, time.Now()); err != nil {
		return err
	}
	s.syncChallengeCountCache(ctx, challengeID, count)
	return nil
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

func (s *challengeService) AddPrize(ctx context.Context, creatorID, challengeID uuid.UUID, prize *models.Prize) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	// Authorization check
	if challenge.CreatorID != creatorID {
		user, err := s.userRepo.GetByID(ctx, creatorID)
		if err != nil || user.Role != models.RoleAdmin {
			return domain.ErrUnauthorized
		}
	}

	// Status check: Only draft challenges can have prizes added
	if challenge.Status != models.ChallengeStatusDraft {
		return fmt.Errorf("prizes can only be added to draft challenges: %w", domain.ErrBadRequest)
	}

	prize.ID = uuid.New()
	prize.ChallengeID = challengeID
	prize.CreatedAt = time.Now()
	prize.UpdatedAt = time.Now()

	return s.prizeRepo.Create(ctx, prize)
}

func (s *challengeService) UpdatePrize(ctx context.Context, creatorID, challengeID, prizeID uuid.UUID, updatedPrize *models.Prize) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	// Authorization check
	if challenge.CreatorID != creatorID {
		user, err := s.userRepo.GetByID(ctx, creatorID)
		if err != nil || user.Role != models.RoleAdmin {
			return domain.ErrUnauthorized
		}
	}

	// Status check
	if challenge.Status != models.ChallengeStatusDraft {
		return fmt.Errorf("prizes can only be modified in draft challenges: %w", domain.ErrBadRequest)
	}

	existingPrize, err := s.prizeRepo.GetByID(ctx, prizeID)
	if err != nil {
		return err
	}

	if existingPrize.ChallengeID != challengeID {
		return fmt.Errorf("prize does not belong to this challenge: %w", domain.ErrBadRequest)
	}

	existingPrize.Title = updatedPrize.Title
	existingPrize.Description = updatedPrize.Description
	existingPrize.ImageURL = updatedPrize.ImageURL
	existingPrize.RankRequired = updatedPrize.RankRequired
	existingPrize.UpdatedAt = time.Now()

	return s.prizeRepo.Update(ctx, existingPrize)
}

func (s *challengeService) DeletePrize(ctx context.Context, creatorID, challengeID, prizeID uuid.UUID) error {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return err
	}

	// Authorization check
	if challenge.CreatorID != creatorID {
		user, err := s.userRepo.GetByID(ctx, creatorID)
		if err != nil || user.Role != models.RoleAdmin {
			return domain.ErrUnauthorized
		}
	}

	// Status check
	if challenge.Status != models.ChallengeStatusDraft {
		return fmt.Errorf("prizes can only be deleted from draft challenges: %w", domain.ErrBadRequest)
	}

	existingPrize, err := s.prizeRepo.GetByID(ctx, prizeID)
	if err != nil {
		return err
	}

	if existingPrize.ChallengeID != challengeID {
		return fmt.Errorf("prize does not belong to this challenge: %w", domain.ErrBadRequest)
	}

	return s.prizeRepo.Delete(ctx, prizeID)
}

func (s *challengeService) GetPrizesByChallengeID(ctx context.Context, challengeID uuid.UUID) ([]*models.Prize, error) {
	return s.prizeRepo.GetByChallengeID(ctx, challengeID)
}

func (s *challengeService) enqueuePublishNotifications(ctx context.Context, challenge *models.Challenge) {
	if s.sqsClient == nil {
		return
	}

	queueURL, err := s.sqsClient.GetQueueURL(ctx, challengeEventsQueueName)
	if err != nil {
		s.logger.Warn("failed to resolve queue for publish notifications", "queue", challengeEventsQueueName, "error", err)
		return
	}

	recipients := map[uuid.UUID]struct{}{
		challenge.CreatorID: {},
	}

	participants, err := s.participationRepo.ListByChallenge(ctx, challenge.ID)
	if err != nil {
		s.logger.Warn("failed to list participants for publish notifications", "challenge_id", challenge.ID, "error", err)
	} else {
		for _, p := range participants {
			recipients[p.UserID] = struct{}{}
		}
	}

	for userID := range recipients {
		user, userErr := s.userRepo.GetByID(ctx, userID)
		if userErr != nil {
			s.logger.Warn("failed to resolve notification recipient", "user_id", userID, "error", userErr)
			continue
		}
		if strings.TrimSpace(user.Email) == "" {
			s.logger.Warn("skipping notification for empty recipient email", "user_id", userID)
			continue
		}

		payload := map[string]any{
			"schema_version":  "v1",
			"type":            "send_email",
			"event_type":      "send_email",
			"producer":        "challenge_service.publish",
			"idempotency_key": fmt.Sprintf("challenge_publish:%s:user:%s", challenge.ID.String(), userID.String()),
			"challenge_id":    challenge.ID.String(),
			"user_id":         userID.String(),
			"to":              user.Email,
			"subject":         fmt.Sprintf("Challenge published: %s", challenge.Title),
			"body":            fmt.Sprintf("<h1>%s</h1><p>Your challenge is now live in Community Fitness Challenge.</p>", challenge.Title),
			"sender":          notificationSenderEmail,
		}

		bodyBytes, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			s.logger.Warn("failed to marshal publish notification payload", "challenge_id", challenge.ID, "user_id", userID, "error", marshalErr)
			continue
		}

		if _, sendErr := s.sqsClient.SendMessage(ctx, queueURL, string(bodyBytes)); sendErr != nil {
			s.logger.Warn("failed to enqueue publish notification", "challenge_id", challenge.ID, "user_id", userID, "error", sendErr)
			continue
		}
	}
}

func (s *challengeService) syncChallengeCountCache(ctx context.Context, challengeID uuid.UUID, count int) {
	counterKey := fmt.Sprintf(domain.RedisKeyChallengeCount, challengeID.String())
	if err := s.redisClient.Set(ctx, counterKey, count, 0).Err(); err != nil {
		s.logger.Warn("Redis error syncing challenge counter", "challenge_id", challengeID, "count", count, "error", err)
	}
}
