package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChallengeRepo struct {
	challenge *models.Challenge
	err       error
}

func (f *fakeChallengeRepo) Create(context.Context, *models.Challenge) error { return nil }
func (f *fakeChallengeRepo) GetByID(context.Context, uuid.UUID) (*models.Challenge, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.challenge, nil
}
func (f *fakeChallengeRepo) GetByIDForUpdate(context.Context, uuid.UUID) (*models.Challenge, error) {
	return f.GetByID(context.Background(), uuid.Nil)
}
func (f *fakeChallengeRepo) Update(context.Context, *models.Challenge) error { return nil }
func (f *fakeChallengeRepo) Delete(context.Context, uuid.UUID) error         { return nil }
func (f *fakeChallengeRepo) List(context.Context, *models.ChallengeStatus) ([]*models.Challenge, error) {
	return nil, nil
}

type fakeParticipationRepo struct {
	mu           sync.Mutex
	participants map[string]*models.Participation
}

func newFakeParticipationRepo() *fakeParticipationRepo {
	return &fakeParticipationRepo{participants: map[string]*models.Participation{}}
}

func participationKey(userID, challengeID uuid.UUID) string {
	return userID.String() + ":" + challengeID.String()
}

func (f *fakeParticipationRepo) Add(_ context.Context, p *models.Participation) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.participants[participationKey(p.UserID, p.ChallengeID)] = p
	return nil
}
func (f *fakeParticipationRepo) Remove(_ context.Context, userID, challengeID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.participants, participationKey(userID, challengeID))
	return nil
}
func (f *fakeParticipationRepo) Get(_ context.Context, userID, challengeID uuid.UUID) (*models.Participation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p, ok := f.participants[participationKey(userID, challengeID)]; ok {
		cp := *p
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}
func (f *fakeParticipationRepo) UpdateCurrentScore(_ context.Context, userID, challengeID uuid.UUID, score int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := participationKey(userID, challengeID)
	p, ok := f.participants[key]
	if !ok {
		return domain.ErrNotFound
	}
	p.CurrentScore = score
	return nil
}
func (f *fakeParticipationRepo) GetParticipantsCount(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (f *fakeParticipationRepo) ListByChallenge(context.Context, uuid.UUID) ([]*models.Participation, error) {
	return nil, nil
}

type fakeDailyLogRepo struct {
	mu        sync.Mutex
	logs      map[string]*models.DailyLog
	createErr error
}

func newFakeDailyLogRepo() *fakeDailyLogRepo {
	return &fakeDailyLogRepo{logs: map[string]*models.DailyLog{}}
}

func logKey(userID, challengeID uuid.UUID, date time.Time) string {
	return userID.String() + ":" + challengeID.String() + ":" + date.UTC().Format("2006-01-02")
}

func (f *fakeDailyLogRepo) Create(_ context.Context, log *models.DailyLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.createErr != nil {
		return f.createErr
	}
	key := logKey(log.UserID, log.ChallengeID, log.LogDate)
	if _, exists := f.logs[key]; exists {
		return domain.ErrAlreadyExists
	}
	cp := *log
	f.logs[key] = &cp
	return nil
}
func (f *fakeDailyLogRepo) GetByID(_ context.Context, id uuid.UUID) (*models.DailyLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, l := range f.logs {
		if l.ID == id {
			cp := *l
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (f *fakeDailyLogRepo) GetByUserChallengeDate(_ context.Context, userID, challengeID uuid.UUID, logDate time.Time) (*models.DailyLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if l, ok := f.logs[logKey(userID, challengeID, logDate)]; ok {
		cp := *l
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}
func (f *fakeDailyLogRepo) ListByUserAndChallenge(_ context.Context, userID, challengeID uuid.UUID) ([]*models.DailyLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*models.DailyLog
	for _, l := range f.logs {
		if l.UserID == userID && l.ChallengeID == challengeID {
			cp := *l
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LogDate.After(out[j].LogDate)
	})
	return out, nil
}

type fakeQueuePublisher struct {
	queueURL string
	messages []string
	errGet   error
	errSend  error
}

func (f *fakeQueuePublisher) GetQueueURL(context.Context, string) (string, error) {
	if f.errGet != nil {
		return "", f.errGet
	}
	if f.queueURL == "" {
		f.queueURL = "queue://test"
	}
	return f.queueURL, nil
}

func (f *fakeQueuePublisher) SendMessage(_ context.Context, _ string, body string) (string, error) {
	if f.errSend != nil {
		return "", f.errSend
	}
	f.messages = append(f.messages, body)
	return uuid.NewString(), nil
}

type fakeRedisClient struct {
	mu     sync.Mutex
	values map[string]string
	zsets  map[string]map[string]float64
}

func newFakeRedisClient() *fakeRedisClient {
	return &fakeRedisClient{
		values: map[string]string{},
		zsets:  map[string]map[string]float64{},
	}
}

func (f *fakeRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewStringCmd(ctx)
	if v, ok := f.values[key]; ok {
		cmd.SetVal(v)
		return cmd
	}
	cmd.SetErr(redis.Nil)
	return cmd
}

func (f *fakeRedisClient) Set(ctx context.Context, key string, value interface{}, _ time.Duration) *redis.StatusCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.values[key] = toString(value)
	cmd := redis.NewStatusCmd(ctx)
	cmd.SetVal("OK")
	return cmd
}

func (f *fakeRedisClient) SetNX(ctx context.Context, key string, value interface{}, _ time.Duration) *redis.BoolCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewBoolCmd(ctx)
	if _, exists := f.values[key]; exists {
		cmd.SetVal(false)
		return cmd
	}
	f.values[key] = toString(value)
	cmd.SetVal(true)
	return cmd
}

func (f *fakeRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewIntCmd(ctx)
	var curr int64
	_, _ = fmt.Sscanf(f.values[key], "%d", &curr)
	curr++
	f.values[key] = intToString(curr)
	cmd.SetVal(curr)
	return cmd
}

func (f *fakeRedisClient) Decr(ctx context.Context, key string) *redis.IntCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewIntCmd(ctx)
	var curr int64
	_, _ = fmt.Sscanf(f.values[key], "%d", &curr)
	curr--
	f.values[key] = intToString(curr)
	cmd.SetVal(curr)
	return cmd
}

func (f *fakeRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewIntCmd(ctx)
	var removed int64
	for _, k := range keys {
		if _, exists := f.values[k]; exists {
			delete(f.values, k)
			removed++
		}
	}
	cmd.SetVal(removed)
	return cmd
}

func (f *fakeRedisClient) Pipeline() redis.Pipeliner { return nil }
func (f *fakeRedisClient) Close() error              { return nil }

func (f *fakeRedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewIntCmd(ctx)
	if _, ok := f.zsets[key]; !ok {
		f.zsets[key] = map[string]float64{}
	}
	added := int64(0)
	for _, member := range members {
		memberID, ok := member.Member.(string)
		if !ok {
			continue
		}
		if _, exists := f.zsets[key][memberID]; !exists {
			added++
		}
		f.zsets[key][memberID] = member.Score
	}
	cmd.SetVal(added)
	return cmd
}

func (f *fakeRedisClient) has(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.values[key]
	return ok
}

func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case int:
		return intToString(int64(x))
	case int64:
		return intToString(x)
	default:
		return ""
	}
}

func intToString(v int64) string { return fmt.Sprintf("%d", v) }

func newLogServiceForTests(challenge *models.Challenge, participationRepo *fakeParticipationRepo, dailyRepo *fakeDailyLogRepo, redisClient *fakeRedisClient, queue *fakeQueuePublisher) LogService {
	return NewLogService(
		&fakeChallengeRepo{challenge: challenge},
		participationRepo,
		dailyRepo,
		redisClient,
		NewScoringService(),
		queue,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func TestLogService_SubmitDailyLog_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}

	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	dRepo := newFakeDailyLogRepo()
	rClient := newFakeRedisClient()
	queue := &fakeQueuePublisher{}
	svc := newLogServiceForTests(challenge, pRepo, dRepo, rClient, queue)

	log, err := svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{
		Steps:         12000,
		Calories:      650,
		ActiveMinutes: 45,
	})
	require.NoError(t, err)
	require.NotNil(t, log)
	assert.Equal(t, 72.50, log.Score)
	assert.Len(t, queue.messages, 1)
}

func TestLogService_SubmitDailyLog_DuplicateRejected(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}

	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	dRepo := newFakeDailyLogRepo()
	rClient := newFakeRedisClient()
	queue := &fakeQueuePublisher{}
	svc := newLogServiceForTests(challenge, pRepo, dRepo, rClient, queue)
	date := time.Now().UTC()

	_, err := svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{LogDate: &date, Steps: 12000, Calories: 650, ActiveMinutes: 45})
	require.NoError(t, err)
	_, err = svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{LogDate: &date, Steps: 12000, Calories: 650, ActiveMinutes: 45})
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestLogService_SubmitDailyLog_FutureDateRejected(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))

	svc := newLogServiceForTests(challenge, pRepo, newFakeDailyLogRepo(), newFakeRedisClient(), &fakeQueuePublisher{})
	future := time.Now().UTC().AddDate(0, 0, 1)

	_, err := svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{LogDate: &future, Steps: 1})
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestLogService_SubmitDailyLog_OutsideChallengeWindowRejected(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	svc := newLogServiceForTests(challenge, pRepo, newFakeDailyLogRepo(), newFakeRedisClient(), &fakeQueuePublisher{})
	oldDate := time.Now().UTC().AddDate(0, 0, -4)

	_, err := svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{LogDate: &oldDate, Steps: 1})
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestLogService_SubmitDailyLog_LockCleanupOnCreateFailure(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	dRepo := newFakeDailyLogRepo()
	dRepo.createErr = errors.New("db write failed")
	rClient := newFakeRedisClient()
	svc := newLogServiceForTests(challenge, pRepo, dRepo, rClient, &fakeQueuePublisher{})

	date := time.Now().UTC()
	_, err := svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{
		LogDate:       &date,
		Steps:         1000,
		Calories:      100,
		ActiveMinutes: 10,
	})
	assert.Error(t, err)

	lockKey := fmt.Sprintf(domain.RedisKeyDailyLogLock, challengeID.String(), userID.String(), normalizeUTCDate(date).Format("2006-01-02"))
	assert.False(t, rClient.has(lockKey))
}

func TestLogService_SubmitDailyLog_TimezoneNormalized(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-72 * time.Hour),
		EndDate:   time.Now().Add(72 * time.Hour),
	}
	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	dRepo := newFakeDailyLogRepo()
	svc := newLogServiceForTests(challenge, pRepo, dRepo, newFakeRedisClient(), &fakeQueuePublisher{})

	utcPlusTwo := time.Date(2026, 4, 15, 23, 30, 0, 0, time.FixedZone("UTC+2", 2*60*60))
	_, err := svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{
		LogDate:       &utcPlusTwo,
		Steps:         3000,
		Calories:      200,
		ActiveMinutes: 20,
	})
	require.NoError(t, err)

	logs, err := dRepo.ListByUserAndChallenge(ctx, userID, challengeID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "2026-04-15", logs[0].LogDate.UTC().Format("2006-01-02"))
}

func TestLogService_SubmitDailyLog_ConcurrentSameDay(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	svc := newLogServiceForTests(challenge, pRepo, newFakeDailyLogRepo(), newFakeRedisClient(), &fakeQueuePublisher{})

	date := time.Now().UTC()
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{
				LogDate:       &date,
				Steps:         12000,
				Calories:      650,
				ActiveMinutes: 45,
			})
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successCount := 0
	alreadyExistsCount := 0
	for err := range results {
		if err == nil {
			successCount++
			continue
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			alreadyExistsCount++
		}
	}
	assert.Equal(t, 1, successCount)
	assert.Equal(t, 1, alreadyExistsCount)
}

func TestLogService_GetDailyLogsWithAggregation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: userID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-10 * 24 * time.Hour),
		EndDate:   time.Now().Add(10 * 24 * time.Hour),
	}
	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	dRepo := newFakeDailyLogRepo()
	svc := newLogServiceForTests(challenge, pRepo, dRepo, newFakeRedisClient(), &fakeQueuePublisher{})

	d0 := time.Now().UTC()
	d1 := d0.AddDate(0, 0, -1)
	_, _ = svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{LogDate: &d0, Steps: 12000, Calories: 650, ActiveMinutes: 45})
	_, _ = svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{LogDate: &d1, Steps: 12000, Calories: 650, ActiveMinutes: 45})

	resp, err := svc.GetDailyLogsWithAggregation(ctx, userID, challengeID)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Aggregation.DaysLogged)
	assert.Equal(t, 2, resp.Aggregation.Streak)
	assert.Equal(t, 145.0, resp.Aggregation.TotalScore)
	assert.Equal(t, 1300, resp.Aggregation.TotalCalories)
}

func TestLogService_GetMyProgress(t *testing.T) {
	ctx := context.Background()
	creatorID := uuid.New()
	userID := uuid.New()
	challengeID := uuid.New()
	challenge := &models.Challenge{
		ID:        challengeID,
		CreatorID: creatorID,
		Status:    models.ChallengeStatusUpcoming,
		StartDate: time.Now().Add(-10 * 24 * time.Hour),
		EndDate:   time.Now().Add(10 * 24 * time.Hour),
	}
	pRepo := newFakeParticipationRepo()
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: creatorID, ChallengeID: challengeID}))
	require.NoError(t, pRepo.Add(ctx, &models.Participation{UserID: userID, ChallengeID: challengeID}))
	svc := newLogServiceForTests(challenge, pRepo, newFakeDailyLogRepo(), newFakeRedisClient(), &fakeQueuePublisher{})

	d := time.Now().UTC()
	_, _ = svc.SubmitDailyLog(ctx, creatorID, challengeID, SubmitDailyLogInput{LogDate: &d, Steps: 12000, Calories: 650, ActiveMinutes: 45})
	_, _ = svc.SubmitDailyLog(ctx, userID, challengeID, SubmitDailyLogInput{LogDate: &d, Steps: 12000, Calories: 650, ActiveMinutes: 45})

	progress, err := svc.GetMyProgress(ctx, userID, challengeID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, progress.RelativeToCreator.Percentage)
	assert.Equal(t, 72.5, progress.Aggregation.TotalScore)
}
