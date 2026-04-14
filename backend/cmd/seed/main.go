package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HeadTDev/fitchallenge/internal/adapter/postgres"
	"github.com/HeadTDev/fitchallenge/internal/config"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/google/uuid"
)

func main() {
	cfg := config.LoadConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := postgres.NewConnection(ctx, cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	userRepo := postgres.NewUserRepo(dbPool)
	challengeRepo := postgres.NewChallengeRepo(dbPool)
	participationRepo := postgres.NewParticipationRepo(dbPool)
	prizeRepo := postgres.NewPrizeRepo(dbPool)

	slog.Info("🌱 Starting database seeding...")

	// 1. Create Users
	users := []models.User{
		{
			ID:          uuid.New(),
			Email:       "admin@fitchallenge.com",
			DisplayName: strPtr("Admin User"),
			Role:        models.RoleAdmin,
			Timezone:    "UTC",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Email:       "creator1@fitchallenge.com",
			DisplayName: strPtr("John Creator"),
			Bio:         strPtr("Fitness enthusiast and challenge creator."),
			Role:        models.RoleCreator,
			Timezone:    "Europe/Budapest",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Email:       "user1@fitchallenge.com",
			DisplayName: strPtr("Alice Runner"),
			Role:        models.RoleParticipant,
			Timezone:    "Europe/Budapest",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Email:       "user2@fitchallenge.com",
			DisplayName: strPtr("Bob Walker"),
			Role:        models.RoleParticipant,
			Timezone:    "UTC",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Email:       "user3@fitchallenge.com",
			DisplayName: strPtr("Charlie Cyclist"),
			Role:        models.RoleParticipant,
			Timezone:    "America/New_York",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for i := range users {
		if err := userRepo.Create(ctx, &users[i]); err != nil {
			slog.Error("Failed to create user", "email", users[i].Email, "error", err)
			continue
		}
		slog.Info("Created user", "email", users[i].Email)
	}

	// 2. Create Challenges
	creatorID := users[1].ID // John Creator
	challenges := []models.Challenge{
		{
			ID:              uuid.New(),
			CreatorID:       creatorID,
			Title:           "Spring Step Challenge",
			Description:     strPtr("Walk 100,000 steps this month!"),
			StartDate:       time.Now().AddDate(0, 0, -5),
			EndDate:         time.Now().AddDate(0, 1, 0),
			Status:          models.ChallengeStatusActive,
			Type:            models.ChallengeTypeSteps,
			Goal:            100000,
			MaxParticipants: 50,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			CreatorID:       creatorID,
			Title:           "Calorie Burner 3000",
			Description:     strPtr("Burn 5,000 calories in two weeks."),
			StartDate:       time.Now().AddDate(0, 0, 2),
			EndDate:         time.Now().AddDate(0, 0, 16),
			Status:          models.ChallengeStatusUpcoming,
			Type:            models.ChallengeTypeCalories,
			Goal:            5000,
			MaxParticipants: 20,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New(),
			CreatorID:       creatorID,
			Title:           "Active Minutes Marathon",
			Description:     strPtr("Get 500 active minutes."),
			StartDate:       time.Now().AddDate(0, 0, -20),
			EndDate:         time.Now().AddDate(0, 0, -1),
			Status:          models.ChallengeStatusFinished,
			Type:            models.ChallengeTypeActiveMinutes,
			Goal:            500,
			MaxParticipants: 100,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for i := range challenges {
		if err := challengeRepo.Create(ctx, &challenges[i]); err != nil {
			slog.Error("Failed to create challenge", "title", challenges[i].Title, "error", err)
			continue
		}
		slog.Info("Created challenge", "title", challenges[i].Title)

		// 3. Create Prizes for each challenge
		prizes := []models.Prize{
			{
				ID:           uuid.New(),
				ChallengeID:  challenges[i].ID,
				Title:        "Gold Medal",
				Description:  strPtr("Awarded to the top performer."),
				RankRequired: 1,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			{
				ID:           uuid.New(),
				ChallengeID:  challenges[i].ID,
				Title:        "Silver Medal",
				Description:  strPtr("Awarded to the second best."),
				RankRequired: 2,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		}

		for j := range prizes {
			if err := prizeRepo.Create(ctx, &prizes[j]); err != nil {
				slog.Error("Failed to create prize", "title", prizes[j].Title, "error", err)
				continue
			}
		}
		slog.Info("Created prizes for challenge", "title", challenges[i].Title)

		// 4. Create Participations
		// Add some users to each challenge
		for _, user := range users {
			// Skip creator joining their own challenge for simplicity in this seed, 
			// though they can in theory.
			if user.ID == creatorID && i == 0 { continue } 

			participation := models.Participation{
				ID:           uuid.New(),
				UserID:       user.ID,
				ChallengeID:  challenges[i].ID,
				CurrentScore: 0,
				Rank:         0,
				JoinedAt:     time.Now(),
				UpdatedAt:    time.Now(),
			}

			if err := participationRepo.Add(ctx, &participation); err != nil {
				slog.Error("Failed to add participation", "user", user.Email, "challenge", challenges[i].Title, "error", err)
				continue
			}
			
			// Update participant count in challenge
			challenges[i].ParticipantCount++
		}
		
		// Update challenge with new participant count
		if err := challengeRepo.Update(ctx, &challenges[i]); err != nil {
			slog.Error("Failed to update challenge participant count", "title", challenges[i].Title, "error", err)
		}
		slog.Info("Added participants to challenge", "title", challenges[i].Title, "count", challenges[i].ParticipantCount)
	}

	slog.Info("✅ Database seeding completed successfully!")
}

func strPtr(s string) *string {
	return &s
}
