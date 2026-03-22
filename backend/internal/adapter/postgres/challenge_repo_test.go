package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeRepo(t *testing.T) {
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Skipping ChallengeRepo test: DB_HOST not set")
	}

	ctx := context.Background()
	cfg := config.LoadConfig()
	pool, err := NewConnection(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewChallengeRepo(pool)

	t.Run("Create and GetByID", func(t *testing.T) {
		id := uuid.New()
		desc := "Test Description"
		challenge := &models.Challenge{
			ID:          id,
			Title:       "Test Challenge",
			Description: &desc,
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(48 * time.Hour),
			Status:      models.ChallengeStatusUpcoming,
			Type:        models.ChallengeTypeSteps,
			Goal:        10000,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, challenge)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, challenge.Title, fetched.Title)
		assert.Equal(t, challenge.Goal, fetched.Goal)
	})

	t.Run("Update", func(t *testing.T) {
		id := uuid.New()
		challenge := &models.Challenge{
			ID:        id,
			Title:     "Old Title",
			StartDate: time.Now(),
			EndDate:   time.Now().Add(time.Hour),
			Status:    models.ChallengeStatusUpcoming,
			Type:      models.ChallengeTypeMixed,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, challenge))

		challenge.Title = "New Title"
		challenge.Status = models.ChallengeStatusActive
		err := repo.Update(ctx, challenge)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "New Title", fetched.Title)
		assert.Equal(t, models.ChallengeStatusActive, fetched.Status)
	})

	t.Run("Delete (Soft Delete)", func(t *testing.T) {
		id := uuid.New()
		challenge := &models.Challenge{
			ID:        id,
			Title:     "To Delete",
			StartDate: time.Now(),
			EndDate:   time.Now().Add(time.Hour),
			Status:    models.ChallengeStatusUpcoming,
			Type:      models.ChallengeTypeMixed,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, challenge))

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("List with Status Filter", func(t *testing.T) {
		status := models.ChallengeStatusFinished
		id := uuid.New()
		challenge := &models.Challenge{
			ID:        id,
			Title:     "Finished Challenge",
			StartDate: time.Now().Add(-48 * time.Hour),
			EndDate:   time.Now().Add(-24 * time.Hour),
			Status:    status,
			Type:      models.ChallengeTypeSteps,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, challenge))

		challenges, err := repo.List(ctx, &status)
		assert.NoError(t, err)
		assert.NotEmpty(t, challenges)
		
		found := false
		for _, c := range challenges {
			if c.ID == id {
				found = true
			}
			assert.Equal(t, status, c.Status)
		}
		assert.True(t, found)
	})
}
