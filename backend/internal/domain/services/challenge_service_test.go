package services

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type mockDBPool struct {
	mock.Mock
}

func (m *mockDBPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	a := m.Called(ctx, sql, args)
	return a.Get(0).(pgconn.CommandTag), a.Error(1)
}
func (m *mockDBPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	a := m.Called(ctx, sql, args)
	return a.Get(0).(pgx.Rows), a.Error(1)
}
func (m *mockDBPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	a := m.Called(ctx, sql, args)
	return a.Get(0).(pgx.Row)
}
func (m *mockDBPool) Begin(ctx context.Context) (pgx.Tx, error) {
	a := m.Called(ctx)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(pgx.Tx), a.Error(1)
}
func (m *mockDBPool) Close() { m.Called() }

type mockTx struct {
	mock.Mock
	pgx.Tx
}

func (m *mockTx) Commit(ctx context.Context) error   { return m.Called(ctx).Error(0) }
func (m *mockTx) Rollback(ctx context.Context) error { return m.Called(ctx).Error(0) }

type mockChallengeRepo struct {
	mock.Mock
}

func (m *mockChallengeRepo) Create(ctx context.Context, c *models.Challenge) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockChallengeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.Challenge), a.Error(1)
}
func (m *mockChallengeRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.Challenge), a.Error(1)
}
func (m *mockChallengeRepo) Update(ctx context.Context, c *models.Challenge) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockChallengeRepo) UpdateParticipantCount(ctx context.Context, id uuid.UUID, count int, updatedAt time.Time) error {
	return m.Called(ctx, id, count, updatedAt).Error(0)
}
func (m *mockChallengeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockChallengeRepo) List(ctx context.Context, s *models.ChallengeStatus) ([]*models.Challenge, error) {
	a := m.Called(ctx, s)
	return a.Get(0).([]*models.Challenge), a.Error(1)
}
func (m *mockChallengeRepo) WithTx(tx interface{}) repositories.ChallengeRepository {
	return m.Called(tx).Get(0).(repositories.ChallengeRepository)
}

type mockParticipationRepo struct {
	mock.Mock
}

func (m *mockParticipationRepo) Add(ctx context.Context, p *models.Participation) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockParticipationRepo) Remove(ctx context.Context, u, c uuid.UUID) error {
	return m.Called(ctx, u, c).Error(0)
}
func (m *mockParticipationRepo) Get(ctx context.Context, u, c uuid.UUID) (*models.Participation, error) {
	a := m.Called(ctx, u, c)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.Participation), a.Error(1)
}
func (m *mockParticipationRepo) UpdateCurrentScore(ctx context.Context, u, c uuid.UUID, score int) error {
	return m.Called(ctx, u, c, score).Error(0)
}
func (m *mockParticipationRepo) GetParticipantsCount(ctx context.Context, c uuid.UUID) (int, error) {
	a := m.Called(ctx, c)
	return a.Int(0), a.Error(1)
}
func (m *mockParticipationRepo) ListByChallenge(ctx context.Context, c uuid.UUID) ([]*models.Participation, error) {
	a := m.Called(ctx, c)
	return a.Get(0).([]*models.Participation), a.Error(1)
}
func (m *mockParticipationRepo) WithTx(tx interface{}) repositories.ParticipationRepository {
	return m.Called(tx).Get(0).(repositories.ParticipationRepository)
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.User), a.Error(1)
}
func (m *mockUserRepo) Create(ctx context.Context, u *models.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) Update(ctx context.Context, u *models.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) GetByAppleID(ctx context.Context, id string) (*models.User, error) {
	a := m.Called(ctx, id)
	return a.Get(0).(*models.User), a.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, e string) (*models.User, error) {
	a := m.Called(ctx, e)
	return a.Get(0).(*models.User), a.Error(1)
}

type mockPrizeRepo struct {
	mock.Mock
}

func (m *mockPrizeRepo) Create(ctx context.Context, p *models.Prize) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPrizeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Prize, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*models.Prize), a.Error(1)
}
func (m *mockPrizeRepo) GetByChallengeID(ctx context.Context, id uuid.UUID) ([]*models.Prize, error) {
	a := m.Called(ctx, id)
	return a.Get(0).([]*models.Prize), a.Error(1)
}
func (m *mockPrizeRepo) Update(ctx context.Context, p *models.Prize) error {
	return m.Called(ctx, p).Error(0)
}
func (m *mockPrizeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockS3Client struct {
	mock.Mock
}

func (m *mockS3Client) UploadFile(ctx context.Context, b, k string, r io.Reader, t string) (string, error) {
	a := m.Called(ctx, b, k, r, t)
	return a.String(0), a.Error(1)
}
func (m *mockS3Client) GetFileURL(b, k string) string { return m.Called(b, k).String(0) }
func (m *mockS3Client) ListFiles(ctx context.Context, b, p string) ([]string, error) {
	a := m.Called(ctx, b, p)
	return a.Get(0).([]string), a.Error(1)
}
func (m *mockS3Client) DeleteFile(ctx context.Context, b, k string) error {
	return m.Called(ctx, b, k).Error(0)
}

type mockRedisClient struct {
	mock.Mock
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.Called(ctx, key).Get(0).(*redis.StringCmd)
}
func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return m.Called(ctx, key, value, expiration).Get(0).(*redis.StatusCmd)
}
func (m *mockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	return m.Called(ctx, key).Get(0).(*redis.IntCmd)
}
func (m *mockRedisClient) Decr(ctx context.Context, key string) *redis.IntCmd {
	return m.Called(ctx, key).Get(0).(*redis.IntCmd)
}
func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return m.Called(ctx, keys).Get(0).(*redis.IntCmd)
}
func (m *mockRedisClient) Pipeline() redis.Pipeliner {
	return m.Called().Get(0).(redis.Pipeliner)
}
func (m *mockRedisClient) Close() error {
	return m.Called().Error(0)
}

type mockQueuePublisher struct {
	mock.Mock
}

func (m *mockQueuePublisher) GetQueueURL(ctx context.Context, queueName string) (string, error) {
	a := m.Called(ctx, queueName)
	return a.String(0), a.Error(1)
}

func (m *mockQueuePublisher) SendMessage(ctx context.Context, queueURL, body string) (string, error) {
	a := m.Called(ctx, queueURL, body)
	return a.String(0), a.Error(1)
}

// --- Tests ---

func setupService() (*mockDBPool, *mockChallengeRepo, *mockUserRepo, *mockParticipationRepo, *mockPrizeRepo, *mockS3Client, *mockRedisClient, *mockQueuePublisher, ChallengeService, *slog.Logger) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	pool := new(mockDBPool)
	cRepo := new(mockChallengeRepo)
	uRepo := new(mockUserRepo)
	pRepo := new(mockParticipationRepo)
	prRepo := new(mockPrizeRepo)
	s3 := new(mockS3Client)
	redisMock := new(mockRedisClient)
	queue := new(mockQueuePublisher)

	service := NewChallengeService(pool, cRepo, uRepo, pRepo, prRepo, s3, redisMock, queue, logger)
	return pool, cRepo, uRepo, pRepo, prRepo, s3, redisMock, queue, service, logger
}

func TestJoinChallenge(t *testing.T) {
	ctx := context.Background()
	pool, cRepo, _, pRepo, _, _, redisMock, _, service, _ := setupService()

	userID := uuid.New()
	challengeID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		challenge := &models.Challenge{
			ID:               challengeID,
			Status:           models.ChallengeStatusUpcoming,
			MaxParticipants:  10,
			ParticipantCount: 5,
		}

		tx := new(mockTx)
		pool.On("Begin", ctx).Return(tx, nil).Once()
		tx.On("Rollback", ctx).Return(nil).Maybe()

		cRepoTx := new(mockChallengeRepo)
		pRepoTx := new(mockParticipationRepo)
		cRepo.On("WithTx", tx).Return(cRepoTx).Once()
		pRepo.On("WithTx", tx).Return(pRepoTx).Once()

		cRepoTx.On("GetByIDForUpdate", ctx, challengeID).Return(challenge, nil).Once()
		pRepoTx.On("Get", ctx, userID, challengeID).Return(nil, domain.ErrNotFound).Once()
		pRepoTx.On("GetParticipantsCount", ctx, challengeID).Return(5, nil).Once()
		pRepoTx.On("Add", ctx, mock.AnythingOfType("*models.Participation")).Return(nil).Once()
		pRepoTx.On("GetParticipantsCount", ctx, challengeID).Return(6, nil).Once()
		cRepoTx.On("UpdateParticipantCount", ctx, challengeID, 6, mock.AnythingOfType("time.Time")).Return(nil).Once()

		tx.On("Commit", ctx).Return(nil).Once()
		redisMock.On("Set", ctx, mock.Anything, 6, time.Duration(0)).Return(redis.NewStatusCmd(ctx)).Once()

		err := service.JoinChallenge(ctx, userID, challengeID)
		assert.NoError(t, err)
	})

	t.Run("Challenge Full", func(t *testing.T) {
		challenge := &models.Challenge{
			ID:               challengeID,
			Status:           models.ChallengeStatusUpcoming,
			MaxParticipants:  5,
			ParticipantCount: 5,
		}

		tx := new(mockTx)
		pool.On("Begin", ctx).Return(tx, nil).Once()
		tx.On("Rollback", ctx).Return(nil).Maybe()

		cRepoTx := new(mockChallengeRepo)
		pRepoTx := new(mockParticipationRepo)
		cRepo.On("WithTx", tx).Return(cRepoTx).Once()
		pRepo.On("WithTx", tx).Return(pRepoTx).Once()

		cRepoTx.On("GetByIDForUpdate", ctx, challengeID).Return(challenge, nil).Once()
		pRepoTx.On("Get", ctx, userID, challengeID).Return(nil, domain.ErrNotFound).Once()
		pRepoTx.On("GetParticipantsCount", ctx, challengeID).Return(5, nil).Once()

		err := service.JoinChallenge(ctx, userID, challengeID)
		assert.ErrorIs(t, err, domain.ErrChallengeFull)
	})
}

func TestLeaveChallenge(t *testing.T) {
	ctx := context.Background()
	pool, cRepo, _, pRepo, _, _, redisMock, _, service, _ := setupService()

	userID := uuid.New()
	challengeID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		tx := new(mockTx)
		pool.On("Begin", ctx).Return(tx, nil).Once()
		tx.On("Rollback", ctx).Return(nil).Maybe()

		cRepoTx := new(mockChallengeRepo)
		pRepoTx := new(mockParticipationRepo)
		cRepo.On("WithTx", tx).Return(cRepoTx).Once()
		pRepo.On("WithTx", tx).Return(pRepoTx).Once()

		pRepoTx.On("Get", ctx, userID, challengeID).Return(&models.Participation{}, nil).Once()
		cRepoTx.On("GetByIDForUpdate", ctx, challengeID).Return(&models.Challenge{ID: challengeID, ParticipantCount: 5}, nil).Once()
		pRepoTx.On("Remove", ctx, userID, challengeID).Return(nil).Once()
		pRepoTx.On("GetParticipantsCount", ctx, challengeID).Return(4, nil).Once()
		cRepoTx.On("UpdateParticipantCount", ctx, challengeID, 4, mock.AnythingOfType("time.Time")).Return(nil).Once()

		tx.On("Commit", ctx).Return(nil).Once()
		redisMock.On("Set", ctx, mock.Anything, 4, time.Duration(0)).Return(redis.NewStatusCmd(ctx)).Once()
		redisMock.On("Del", ctx, mock.Anything).Return(redis.NewIntCmd(ctx)).Once()

		err := service.LeaveChallenge(ctx, userID, challengeID)
		assert.NoError(t, err)
	})
}

func TestCreateChallenge(t *testing.T) {
	ctx := context.Background()
	_, cRepo, uRepo, _, _, _, _, _, service, _ := setupService()

	creatorID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		uRepo.On("GetByID", ctx, creatorID).Return(&models.User{Role: models.RoleCreator}, nil).Once()
		cRepo.On("Create", ctx, mock.AnythingOfType("*models.Challenge")).Return(nil).Once()

		err := service.CreateChallenge(ctx, creatorID, &models.Challenge{Title: "Test"})
		assert.NoError(t, err)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		uRepo.On("GetByID", ctx, creatorID).Return(&models.User{Role: models.RoleParticipant}, nil).Once()

		err := service.CreateChallenge(ctx, creatorID, &models.Challenge{Title: "Test"})
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}

func TestPublishChallenge(t *testing.T) {
	ctx := context.Background()
	_, cRepo, uRepo, pRepo, _, _, _, queue, service, _ := setupService()

	creatorID := uuid.New()
	challengeID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		challenge := &models.Challenge{ID: challengeID, CreatorID: creatorID, Status: models.ChallengeStatusDraft}
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()
		cRepo.On("Update", ctx, mock.MatchedBy(func(c *models.Challenge) bool {
			return c.Status == models.ChallengeStatusUpcoming
		})).Return(nil).Once()
		pRepo.On("ListByChallenge", ctx, challengeID).Return([]*models.Participation{}, nil).Once()
		uRepo.On("GetByID", ctx, creatorID).Return(&models.User{ID: creatorID, Email: "creator@example.com"}, nil).Once()
		queue.On("GetQueueURL", ctx, challengeEventsQueueName).Return("queue-url", nil).Once()
		queue.On("SendMessage", ctx, "queue-url", mock.MatchedBy(func(body string) bool {
			return strings.Contains(body, "\"send_email\"") && strings.Contains(body, creatorID.String())
		})).Return("msg-id", nil).Once()

		err := service.PublishChallenge(ctx, creatorID, challengeID)
		assert.NoError(t, err)
	})

	t.Run("Not Draft", func(t *testing.T) {
		challenge := &models.Challenge{ID: challengeID, CreatorID: creatorID, Status: models.ChallengeStatusActive}
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()

		err := service.PublishChallenge(ctx, creatorID, challengeID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only draft challenges can be published")
	})
}

func TestAddPrize(t *testing.T) {
	ctx := context.Background()
	_, cRepo, _, _, prRepo, _, _, _, service, _ := setupService()

	creatorID := uuid.New()
	challengeID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		challenge := &models.Challenge{ID: challengeID, CreatorID: creatorID, Status: models.ChallengeStatusDraft}
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()
		prRepo.On("Create", ctx, mock.AnythingOfType("*models.Prize")).Return(nil).Once()

		err := service.AddPrize(ctx, creatorID, challengeID, &models.Prize{Title: "Gold"})
		assert.NoError(t, err)
	})

	t.Run("Forbidden on Published", func(t *testing.T) {
		challenge := &models.Challenge{ID: challengeID, CreatorID: creatorID, Status: models.ChallengeStatusUpcoming}
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()

		err := service.AddPrize(ctx, creatorID, challengeID, &models.Prize{Title: "Gold"})
		assert.ErrorIs(t, err, domain.ErrBadRequest)
	})
}
