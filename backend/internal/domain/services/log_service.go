package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const logEventsQueueName = "fitchallenge-jobs"

type queuePublisher interface {
	GetQueueURL(ctx context.Context, queueName string) (string, error)
	SendMessage(ctx context.Context, queueURL, body string) (string, error)
}

type redisSetNXClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
}

type SubmitDailyLogInput struct {
	LogDate           *time.Time
	Steps             int
	Calories          int
	ActiveMinutes     int
	HealthKitDataHash *string
	SourceBundleIDs   []string
}

type LogService interface {
	SubmitDailyLog(ctx context.Context, userID, challengeID uuid.UUID, input SubmitDailyLogInput) (*models.DailyLog, error)
	GetDailyLogsWithAggregation(ctx context.Context, userID, challengeID uuid.UUID) (*models.DailyLogListResponse, error)
	GetMyProgress(ctx context.Context, userID, challengeID uuid.UUID) (*models.MyProgressResponse, error)
}

type logService struct {
	challengeRepo     repositories.ChallengeRepository
	participationRepo repositories.ParticipationRepository
	dailyLogRepo      repositories.DailyLogRepository
	redisClient       domain.RedisClient
	scoringService    ScoringService
	sqsClient         queuePublisher
	logger            *slog.Logger
}

func NewLogService(
	challengeRepo repositories.ChallengeRepository,
	participationRepo repositories.ParticipationRepository,
	dailyLogRepo repositories.DailyLogRepository,
	redisClient domain.RedisClient,
	scoringService ScoringService,
	sqsClient queuePublisher,
	logger *slog.Logger,
) LogService {
	return &logService{
		challengeRepo:     challengeRepo,
		participationRepo: participationRepo,
		dailyLogRepo:      dailyLogRepo,
		redisClient:       redisClient,
		scoringService:    scoringService,
		sqsClient:         sqsClient,
		logger:            logger,
	}
}

func (s *logService) SubmitDailyLog(ctx context.Context, userID, challengeID uuid.UUID, input SubmitDailyLogInput) (*models.DailyLog, error) {
	if userID == uuid.Nil || challengeID == uuid.Nil {
		return nil, fmt.Errorf("missing user or challenge id: %w", domain.ErrInvalidInput)
	}
	if input.Steps < 0 || input.Calories < 0 || input.ActiveMinutes < 0 {
		return nil, fmt.Errorf("negative metrics are not allowed: %w", domain.ErrInvalidInput)
	}
	if input.Steps == 0 && input.Calories == 0 && input.ActiveMinutes == 0 {
		return nil, fmt.Errorf("at least one metric must be greater than zero: %w", domain.ErrInvalidInput)
	}

	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if challenge.Status == models.ChallengeStatusDraft || challenge.Status == models.ChallengeStatusFinished {
		return nil, fmt.Errorf("daily logs are not accepted in challenge status %s: %w", challenge.Status, domain.ErrBadRequest)
	}

	if _, err := s.participationRepo.Get(ctx, userID, challengeID); err != nil {
		return nil, err
	}

	logDate := normalizeUTCDate(time.Now().UTC())
	if input.LogDate != nil && !input.LogDate.IsZero() {
		logDate = normalizeUTCDate(*input.LogDate)
	}
	if logDate.After(normalizeUTCDate(time.Now().UTC())) {
		return nil, fmt.Errorf("future log date is not allowed: %w", domain.ErrInvalidInput)
	}
	if logDate.Before(normalizeUTCDate(challenge.StartDate)) || logDate.After(normalizeUTCDate(challenge.EndDate)) {
		return nil, fmt.Errorf("log date must be within challenge timeframe: %w", domain.ErrInvalidInput)
	}

	score, err := s.scoringService.Calculate(ScoreInput{
		Steps:         input.Steps,
		Calories:      input.Calories,
		ActiveMinutes: input.ActiveMinutes,
	})
	if err != nil {
		return nil, err
	}

	redisNX, ok := s.redisClient.(redisSetNXClient)
	if !ok {
		return nil, fmt.Errorf("redis client missing SetNX capability: %w", domain.ErrInternal)
	}

	lockKey := fmt.Sprintf(domain.RedisKeyDailyLogLock, challengeID.String(), userID.String(), logDate.Format("2006-01-02"))
	acquired, err := redisNX.SetNX(ctx, lockKey, "1", 48*time.Hour).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to set daily log lock: %w", err)
	}
	if !acquired {
		return nil, domain.ErrAlreadyExists
	}

	sourceBundleIDs := input.SourceBundleIDs
	if sourceBundleIDs == nil {
		sourceBundleIDs = []string{}
	}

	log := &models.DailyLog{
		ID:                uuid.New(),
		UserID:            userID,
		ChallengeID:       challengeID,
		LogDate:           logDate,
		Steps:             input.Steps,
		Calories:          input.Calories,
		ActiveMinutes:     input.ActiveMinutes,
		Score:             score.Score,
		HealthKitDataHash: input.HealthKitDataHash,
		SourceBundleIDs:   sourceBundleIDs,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := s.dailyLogRepo.Create(ctx, log); err != nil {
		if !errors.Is(err, domain.ErrAlreadyExists) {
			if unlockErr := s.redisClient.Del(ctx, lockKey).Err(); unlockErr != nil {
				s.logger.Warn("failed to release daily log lock after create failure", "lock_key", lockKey, "error", unlockErr)
			}
		}
		return nil, err
	}

	if err := s.syncParticipationScore(ctx, userID, challengeID); err != nil {
		s.logger.Warn("failed to sync participation score after daily log submission", "user_id", userID, "challenge_id", challengeID, "error", err)
	}

	s.publishLogSubmittedEvent(ctx, log)

	return log, nil
}

func (s *logService) GetDailyLogsWithAggregation(ctx context.Context, userID, challengeID uuid.UUID) (*models.DailyLogListResponse, error) {
	if userID == uuid.Nil || challengeID == uuid.Nil {
		return nil, fmt.Errorf("missing user or challenge id: %w", domain.ErrInvalidInput)
	}

	if _, err := s.challengeRepo.GetByID(ctx, challengeID); err != nil {
		return nil, err
	}
	if _, err := s.participationRepo.Get(ctx, userID, challengeID); err != nil {
		return nil, err
	}

	logs, err := s.dailyLogRepo.ListByUserAndChallenge(ctx, userID, challengeID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.DailyLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = log.ToResponse()
	}

	return &models.DailyLogListResponse{
		Logs:        responses,
		Aggregation: calculateDailyLogAggregation(logs),
	}, nil
}

func (s *logService) GetMyProgress(ctx context.Context, userID, challengeID uuid.UUID) (*models.MyProgressResponse, error) {
	if userID == uuid.Nil || challengeID == uuid.Nil {
		return nil, fmt.Errorf("missing user or challenge id: %w", domain.ErrInvalidInput)
	}

	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if _, err := s.participationRepo.Get(ctx, userID, challengeID); err != nil {
		return nil, err
	}

	myLogs, err := s.dailyLogRepo.ListByUserAndChallenge(ctx, userID, challengeID)
	if err != nil {
		return nil, err
	}
	myAgg := calculateDailyLogAggregation(myLogs)

	creatorScore := 0.0
	creatorLogs, err := s.dailyLogRepo.ListByUserAndChallenge(ctx, challenge.CreatorID, challengeID)
	if err != nil {
		return nil, err
	}
	if len(creatorLogs) > 0 {
		creatorScore = calculateDailyLogAggregation(creatorLogs).TotalScore
	}

	percentage := 0.0
	if creatorScore > 0 {
		percentage = round2((myAgg.TotalScore / creatorScore) * 100)
	}

	return &models.MyProgressResponse{
		ChallengeID: challengeID,
		Aggregation: myAgg,
		CreatorStats: models.CreatorStats{
			CreatorID: challenge.CreatorID,
			Score:     creatorScore,
		},
		RelativeToCreator: models.RelativeToCreator{
			CreatorScore: creatorScore,
			MyScore:      myAgg.TotalScore,
			Percentage:   percentage,
		},
	}, nil
}

func (s *logService) publishLogSubmittedEvent(ctx context.Context, log *models.DailyLog) {
	queueURL, err := s.sqsClient.GetQueueURL(ctx, logEventsQueueName)
	if err != nil {
		s.logger.Warn("failed to resolve SQS queue URL for log event", "queue", logEventsQueueName, "error", err)
		return
	}

	eventBody, err := json.Marshal(map[string]interface{}{
		"event_type":      "log_submitted",
		"log_id":          log.ID.String(),
		"user_id":         log.UserID.String(),
		"challenge_id":    log.ChallengeID.String(),
		"log_date":        log.LogDate.Format("2006-01-02"),
		"score":           log.Score,
		"submitted_at":    log.CreatedAt.Format(time.RFC3339),
		"scoring_version": ScoringVersionV1,
	})
	if err != nil {
		s.logger.Warn("failed to marshal log submitted event", "log_id", log.ID, "error", err)
		return
	}

	if _, err := s.sqsClient.SendMessage(ctx, queueURL, string(eventBody)); err != nil {
		s.logger.Warn("failed to publish log submitted event", "queue_url", queueURL, "log_id", log.ID, "error", err)
	}
}

func normalizeUTCDate(t time.Time) time.Time {
	year, month, day := t.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func calculateDailyLogAggregation(logs []*models.DailyLog) models.DailyLogAggregation {
	totalScore := 0.0
	totalCalories := 0
	for _, log := range logs {
		totalScore += log.Score
		totalCalories += log.Calories
	}

	return models.DailyLogAggregation{
		TotalScore:    round2(totalScore),
		TotalCalories: totalCalories,
		DaysLogged:    len(logs),
		Streak:        calculateStreak(logs),
	}
}

func calculateStreak(logs []*models.DailyLog) int {
	if len(logs) == 0 {
		return 0
	}

	expected := normalizeUTCDate(logs[0].LogDate)
	streak := 0

	for _, log := range logs {
		logDate := normalizeUTCDate(log.LogDate)
		if !logDate.Equal(expected) {
			break
		}
		streak++
		expected = expected.AddDate(0, 0, -1)
	}

	return streak
}

func (s *logService) syncParticipationScore(ctx context.Context, userID, challengeID uuid.UUID) error {
	logs, err := s.dailyLogRepo.ListByUserAndChallenge(ctx, userID, challengeID)
	if err != nil {
		return err
	}

	aggregation := calculateDailyLogAggregation(logs)
	return s.participationRepo.UpdateCurrentScore(ctx, userID, challengeID, int(round2(aggregation.TotalScore)))
}
