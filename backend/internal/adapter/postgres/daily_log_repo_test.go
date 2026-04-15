package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailyLogRepo(t *testing.T) {
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Skipping DailyLogRepo test: DB_HOST not set")
	}

	ctx := context.Background()
	cfg := config.LoadConfig()
	pool, err := NewConnection(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	userRepo := NewUserRepo(pool)
	challengeRepo := NewChallengeRepo(pool)
	repo := NewDailyLogRepo(pool)

	user := &models.User{
		ID:          uuid.New(),
		Email:       "dailylog-test-" + uuid.NewString() + "@example.com",
		DisplayName: strPtr("DailyLog Tester"),
		Timezone:    "UTC",
		Role:        models.RoleCreator,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, userRepo.Create(ctx, user))

	challenge := &models.Challenge{
		ID:        uuid.New(),
		CreatorID: user.ID,
		Title:     "DailyLog Test Challenge",
		StartDate: time.Now().Add(-time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
		Status:    models.ChallengeStatusActive,
		Type:      models.ChallengeTypeSteps,
		Goal:      10000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, challengeRepo.Create(ctx, challenge))

	logDate := time.Now().UTC()

	t.Run("Create and GetByUserChallengeDate", func(t *testing.T) {
		log := &models.DailyLog{
			ID:            uuid.New(),
			UserID:        user.ID,
			ChallengeID:   challenge.ID,
			LogDate:       logDate,
			Steps:         12000,
			Calories:      650,
			ActiveMinutes: 45,
			Score:         72.50,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.Create(ctx, log)
		require.NoError(t, err)

		got, err := repo.GetByUserChallengeDate(ctx, user.ID, challenge.ID, logDate)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 12000, got.Steps)
		assert.Equal(t, 650, got.Calories)
		assert.Equal(t, 45, got.ActiveMinutes)
	})

	t.Run("Duplicate same day violates unique", func(t *testing.T) {
		dup := &models.DailyLog{
			ID:            uuid.New(),
			UserID:        user.ID,
			ChallengeID:   challenge.ID,
			LogDate:       logDate,
			Steps:         5000,
			Calories:      300,
			ActiveMinutes: 20,
			Score:         40,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.Create(ctx, dup)
		assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	})

	t.Run("Negative metric violates check", func(t *testing.T) {
		invalid := &models.DailyLog{
			ID:            uuid.New(),
			UserID:        user.ID,
			ChallengeID:   challenge.ID,
			LogDate:       logDate.AddDate(0, 0, 1),
			Steps:         -1,
			Calories:      200,
			ActiveMinutes: 10,
			Score:         1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.Create(ctx, invalid)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}

func strPtr(s string) *string {
	return &s
}
