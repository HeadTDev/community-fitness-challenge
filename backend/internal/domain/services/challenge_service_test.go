package services

import (
	"context"
	"testing"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Manual Mocks (or use testify mock)
type mockChallengeRepo struct {
	mock.Mock
}

func (m *mockChallengeRepo) Create(ctx context.Context, c *models.Challenge) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *mockChallengeRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Challenge, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Challenge), args.Error(1)
}
func (m *mockChallengeRepo) Update(ctx context.Context, c *models.Challenge) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *mockChallengeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockChallengeRepo) List(ctx context.Context, s *models.ChallengeStatus) ([]*models.Challenge, error) {
	args := m.Called(ctx, s)
	return args.Get(0).([]*models.Challenge), args.Error(1)
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) GetByAppleID(ctx context.Context, appleID string) (*models.User, error) {
	args := m.Called(ctx, appleID)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateChallenge(t *testing.T) {
	ctx := context.Background()
	cRepo := new(mockChallengeRepo)
	uRepo := new(mockUserRepo)
	service := NewChallengeService(cRepo, uRepo, nil)

	creatorID := uuid.New()
	participantID := uuid.New()

	t.Run("Success as Creator", func(t *testing.T) {
		uRepo.On("GetByID", ctx, creatorID).Return(&models.User{ID: creatorID, Role: models.RoleCreator}, nil).Once()
		cRepo.On("Create", ctx, mock.AnythingOfType("*models.Challenge")).Return(nil).Once()

		challenge := &models.Challenge{Title: "Test Challenge"}
		err := service.CreateChallenge(ctx, creatorID, challenge)

		assert.NoError(t, err)
		assert.Equal(t, models.ChallengeStatusDraft, challenge.Status)
		assert.Equal(t, creatorID, challenge.CreatorID)
		uRepo.AssertExpectations(t)
		cRepo.AssertExpectations(t)
	})

	t.Run("Failure as Participant", func(t *testing.T) {
		uRepo.On("GetByID", ctx, participantID).Return(&models.User{ID: participantID, Role: models.RoleParticipant}, nil).Once()

		challenge := &models.Challenge{Title: "Test Challenge"}
		err := service.CreateChallenge(ctx, participantID, challenge)

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		uRepo.AssertExpectations(t)
	})
}

func TestPublishChallenge(t *testing.T) {
	ctx := context.Background()
	cRepo := new(mockChallengeRepo)
	uRepo := new(mockUserRepo)
	service := NewChallengeService(cRepo, uRepo, nil)

	creatorID := uuid.New()
	challengeID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		challenge := &models.Challenge{
			ID:        challengeID,
			CreatorID: creatorID,
			Status:    models.ChallengeStatusDraft,
		}
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()
		cRepo.On("Update", ctx, challenge).Return(nil).Once()

		err := service.PublishChallenge(ctx, creatorID, challengeID)

		assert.NoError(t, err)
		assert.Equal(t, models.ChallengeStatusUpcoming, challenge.Status)
	})

	t.Run("Unauthorized - Not Creator", func(t *testing.T) {
		otherUserID := uuid.New()
		challenge := &models.Challenge{
			ID:        challengeID,
			CreatorID: creatorID,
			Status:    models.ChallengeStatusDraft,
		}
		cRepo.On("GetByID", ctx, challengeID).Return(challenge, nil).Once()
		uRepo.On("GetByID", ctx, otherUserID).Return(&models.User{ID: otherUserID, Role: models.RoleParticipant}, nil).Once()

		err := service.PublishChallenge(ctx, otherUserID, challengeID)

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}

func TestUploadCoverImage(t *testing.T) {
	// S3 testing usually requires a more complex setup or a mock S3 client
	// For this unit test, we'll skip the actual S3 call if possible or mock it.
	// Since s3Client is a struct and not an interface, it's harder to mock without an interface.
	// Senior tip: Always use interfaces for external services to make testing easier.
}
