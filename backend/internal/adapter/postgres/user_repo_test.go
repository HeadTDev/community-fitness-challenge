package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/adapter/postgres"
	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo(t *testing.T) {
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Skipping integration test: DB_HOST not set")
	}

	ctx := context.Background()
	cfg := config.LoadConfig()
	
	pool, err := postgres.NewConnection(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewUserRepo(pool)

	// Test Data
	userID := uuid.New()
	email := fmt.Sprintf("test-%s@example.com", userID.String()[:8])
	appleID := "apple-" + userID.String()[:8]
	
	user := &models.User{
		ID:          userID,
		Email:       email,
		AppleID:     &appleID,
		Role:        models.RoleParticipant,
		Timezone:    "UTC",
		CreatedAt:   time.Now().Truncate(time.Microsecond),
		UpdatedAt:   time.Now().Truncate(time.Microsecond),
	}

	t.Run("Create User", func(t *testing.T) {
		err := repo.Create(ctx, user)
		assert.NoError(t, err)
	})

	t.Run("Get User By ID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("Get User By Apple ID", func(t *testing.T) {
		found, err := repo.GetByAppleID(ctx, appleID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, userID, found.ID)
	})

	t.Run("Update User", func(t *testing.T) {
		newName := "Updated Name"
		user.DisplayName = &newName
		user.UpdatedAt = time.Now().Truncate(time.Microsecond)
		
		err := repo.Update(ctx, user)
		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, newName, *found.DisplayName)
	})

	t.Run("Delete User (Soft Delete)", func(t *testing.T) {
		err := repo.Delete(ctx, userID)
		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}
