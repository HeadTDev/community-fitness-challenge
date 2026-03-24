package services

import (
	"context"
	"io"
	"testing"
	"time"
	"log/slog"
	"os"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
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
	pgx.Tx // Satisfy pgx.Tx interface
}
func (m *mockTx) Commit(ctx context.Context) error { return m.Called(ctx).Error(0) }
func (m *mockTx) Rollback(ctx context.Context) error { return m.Called(ctx).Error(0) }

type mockChallengeRepo struct {
	mock.Mock
}
func (m *mockChallengeRepo) Create(ctx context.Context, c *models.Challenge) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mockChallengeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil { return nil, a.Error(1) }
	return a.Get(0).(*models.Challenge), a.Error(1)
}
func (m *mockChallengeRepo) Update(ctx context.Context, c *models.Challenge) error {
	return m.Called(ctx, c).Error(0)
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
	if a.Get(0) == nil { return nil, a.Error(1) }
	return a.Get(0).(*models.Participation), a.Error(1)
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
	if a.Get(0) == nil { return nil, a.Error(1) }
	return a.Get(0).(*models.User), a.Error(1)
}
func (m *mockUserRepo) Create(ctx context.Context, u *models.User) error { return m.Called(ctx, u).Error(0) }
func (m *mockUserRepo) Update(ctx context.Context, u *models.User) error { return m.Called(ctx, u).Error(0) }
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error { return m.Called(ctx, id).Error(0) }
func (m *mockUserRepo) GetByAppleID(ctx context.Context, id string) (*models.User, error) {
	a := m.Called(ctx, id)
	return a.Get(0).(*models.User), a.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, e string) (*models.User, error) {
	a := m.Called(ctx, e)
	return a.Get(0).(*models.User), a.Error(1)
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

// --- Tests ---

func TestJoinChallenge(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	
	pool := new(mockDBPool)
	cRepo := new(mockChallengeRepo)
	uRepo := new(mockUserRepo)
	pRepo := new(mockParticipationRepo)
	s3 := new(mockS3Client)
	redisMock := new(mockRedisClient)
	
	service := NewChallengeService(pool, cRepo, uRepo, pRepo, s3, redisMock, logger)

	userID := uuid.New()
	challengeID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		challenge := &models.Challenge{
			ID: challengeID,
			Status: models.ChallengeStatusUpcoming,
			MaxParticipants: 10,
			ParticipantCount: 5,
		}
		
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()
		pRepo.On("Get", ctx, userID, challengeID).Return(nil, domain.ErrNotFound).Once()
		
		tx := new(mockTx)
		pool.On("Begin", ctx).Return(tx, nil).Once()
		tx.On("Rollback", ctx).Return(nil).Maybe()
		
		cRepoTx := new(mockChallengeRepo)
		pRepoTx := new(mockParticipationRepo)
		cRepo.On("WithTx", tx).Return(cRepoTx).Once()
		pRepo.On("WithTx", tx).Return(pRepoTx).Once()
		
		pRepoTx.On("Add", ctx, mock.AnythingOfType("*models.Participation")).Return(nil).Once()
		pRepoTx.On("GetParticipantsCount", ctx, challengeID).Return(6, nil).Once()
		cRepoTx.On("Update", ctx, mock.MatchedBy(func(c *models.Challenge) bool {
			return c.ParticipantCount == 6
		})).Return(nil).Once()
		
		tx.On("Commit", ctx).Return(nil).Once()
		
		redisMock.On("Incr", ctx, mock.Anything).Return(redis.NewIntCmd(ctx)).Once()
		
		err := service.JoinChallenge(ctx, userID, challengeID)
		
		assert.NoError(t, err)
		pool.AssertExpectations(t)
		cRepo.AssertExpectations(t)
		pRepo.AssertExpectations(t)
		tx.AssertExpectations(t)
		redisMock.AssertExpectations(t)
	})

	t.Run("Challenge Full", func(t *testing.T) {
		challenge := &models.Challenge{
			ID: challengeID,
			Status: models.ChallengeStatusUpcoming,
			MaxParticipants: 5,
			ParticipantCount: 5,
		}
		
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()
		pRepo.On("Get", ctx, userID, challengeID).Return(nil, domain.ErrNotFound).Once()
		
		err := service.JoinChallenge(ctx, userID, challengeID)
		
		assert.ErrorIs(t, err, domain.ErrChallengeFull)
	})
}

func TestLeaveChallenge(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	
	pool := new(mockDBPool)
	cRepo := new(mockChallengeRepo)
	uRepo := new(mockUserRepo)
	pRepo := new(mockParticipationRepo)
	s3 := new(mockS3Client)
	redisMock := new(mockRedisClient)
	
	service := NewChallengeService(pool, cRepo, uRepo, pRepo, s3, redisMock, logger)

	userID := uuid.New()
	challengeID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		pRepo.On("Get", ctx, userID, challengeID).Return(&models.Participation{}, nil).Once()
		
		tx := new(mockTx)
		pool.On("Begin", ctx).Return(tx, nil).Once()
		tx.On("Rollback", ctx).Return(nil).Maybe()
		
		cRepoTx := new(mockChallengeRepo)
		pRepoTx := new(mockParticipationRepo)
		cRepo.On("WithTx", tx).Return(cRepoTx).Once()
		pRepo.On("WithTx", tx).Return(pRepoTx).Once()
		
		pRepoTx.On("Remove", ctx, userID, challengeID).Return(nil).Once()
		pRepoTx.On("GetParticipantsCount", ctx, challengeID).Return(4, nil).Once()
		
		challenge := &models.Challenge{ID: challengeID, ParticipantCount: 5}
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()
		cRepoTx.On("Update", ctx, mock.MatchedBy(func(c *models.Challenge) bool {
			return c.ParticipantCount == 4
		})).Return(nil).Once()
		
		tx.On("Commit", ctx).Return(nil).Once()
		redisMock.On("Decr", ctx, mock.Anything).Return(redis.NewIntCmd(ctx)).Once()
		
		err := service.LeaveChallenge(ctx, userID, challengeID)
		
		assert.NoError(t, err)
	})
}
