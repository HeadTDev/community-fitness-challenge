package services

import (
	"context"
	"testing"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubLeaderboardRepo struct {
	top      []models.LeaderboardEntry
	total    int64
	rank     int
	topErr   error
	totalErr error
	rankErr  error

	getTopCalls int
}

func (s *stubLeaderboardRepo) UpdateScore(context.Context, uuid.UUID, uuid.UUID, float64) error {
	return nil
}
func (s *stubLeaderboardRepo) GetRank(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return s.rank, s.rankErr
}
func (s *stubLeaderboardRepo) GetTopN(context.Context, uuid.UUID, int64) ([]models.LeaderboardEntry, error) {
	s.getTopCalls++
	return s.top, s.topErr
}
func (s *stubLeaderboardRepo) GetAroundUser(context.Context, uuid.UUID, uuid.UUID, int64) ([]models.LeaderboardEntry, error) {
	return nil, nil
}
func (s *stubLeaderboardRepo) GetTotalCount(context.Context, uuid.UUID) (int64, error) {
	return s.total, s.totalErr
}

type stubParticipationRepo struct {
	me *models.Participation
}

func (s *stubParticipationRepo) Add(context.Context, *models.Participation) error   { return nil }
func (s *stubParticipationRepo) Remove(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (s *stubParticipationRepo) UpdateCurrentScore(context.Context, uuid.UUID, uuid.UUID, int) error {
	return nil
}
func (s *stubParticipationRepo) GetParticipantsCount(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (s *stubParticipationRepo) ListByChallenge(context.Context, uuid.UUID) ([]*models.Participation, error) {
	return nil, nil
}
func (s *stubParticipationRepo) Get(context.Context, uuid.UUID, uuid.UUID) (*models.Participation, error) {
	return s.me, nil
}

type stubChallengeRepo struct{}

func (s *stubChallengeRepo) Create(context.Context, *models.Challenge) error { return nil }
func (s *stubChallengeRepo) GetByID(context.Context, uuid.UUID) (*models.Challenge, error) {
	return &models.Challenge{}, nil
}
func (s *stubChallengeRepo) GetByIDForUpdate(context.Context, uuid.UUID) (*models.Challenge, error) {
	return &models.Challenge{}, nil
}
func (s *stubChallengeRepo) Update(context.Context, *models.Challenge) error { return nil }
func (s *stubChallengeRepo) UpdateParticipantCount(context.Context, uuid.UUID, int, time.Time) error {
	return nil
}
func (s *stubChallengeRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (s *stubChallengeRepo) List(context.Context, *models.ChallengeStatus) ([]*models.Challenge, error) {
	return nil, nil
}

func TestLeaderboardService_GetAbsoluteLeaderboard_FallsBackOnPartialRedisTop(t *testing.T) {
	ctx := context.Background()
	challengeID := uuid.New()
	userID := uuid.New()

	primary := &stubLeaderboardRepo{
		top: []models.LeaderboardEntry{
			{UserID: userID, Score: 100, Rank: 1},
		},
		total: 3,
		rank:  2,
	}
	fallback := &stubLeaderboardRepo{
		top: []models.LeaderboardEntry{
			{UserID: uuid.New(), Score: 110, Rank: 1},
			{UserID: userID, Score: 100, Rank: 2},
			{UserID: uuid.New(), Score: 95, Rank: 3},
		},
		total: 3,
		rank:  2,
	}
	participationRepo := &stubParticipationRepo{
		me: &models.Participation{
			UserID:       userID,
			ChallengeID:  challengeID,
			CurrentScore: 100,
		},
	}

	service := NewLeaderboardService(primary, fallback, participationRepo, &stubChallengeRepo{}, nil)
	resp, err := service.GetAbsoluteLeaderboard(ctx, challengeID, userID, 3)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Len(t, resp.Top, 3)
	assert.Equal(t, 2, resp.MyPosition.Rank)
	assert.Equal(t, 1, primary.getTopCalls)
	assert.Equal(t, 1, fallback.getTopCalls)
}

func TestShouldUseFallbackForTop(t *testing.T) {
	assert.False(t, shouldUseFallbackForTop([]models.LeaderboardEntry{}, 0, 20, nil, nil))
	assert.True(t, shouldUseFallbackForTop([]models.LeaderboardEntry{{Rank: 1}}, 3, 3, nil, nil))
	assert.False(t, shouldUseFallbackForTop([]models.LeaderboardEntry{{Rank: 1}, {Rank: 2}}, 2, 20, nil, nil))
}
