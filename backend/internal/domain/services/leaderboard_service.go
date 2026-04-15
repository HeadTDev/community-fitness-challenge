package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/HeadTDev/fitchallenge/internal/domain"
	"github.com/HeadTDev/fitchallenge/internal/domain/models"
	"github.com/HeadTDev/fitchallenge/internal/domain/repositories"
	"github.com/google/uuid"
)

type LeaderboardService interface {
	GetAbsoluteLeaderboard(ctx context.Context, challengeID, userID uuid.UUID, limit int64) (*models.AbsoluteLeaderboardResponse, error)
	GetRelativeLeaderboard(ctx context.Context, challengeID, userID uuid.UUID, radius int64) (*models.RelativeLeaderboardResponse, error)
}

type leaderboardService struct {
	leaderboardRepo   repositories.LeaderboardRepository
	participationRepo repositories.ParticipationRepository
	challengeRepo     repositories.ChallengeRepository
}

func NewLeaderboardService(
	leaderboardRepo repositories.LeaderboardRepository,
	participationRepo repositories.ParticipationRepository,
	challengeRepo repositories.ChallengeRepository,
) LeaderboardService {
	return &leaderboardService{
		leaderboardRepo:   leaderboardRepo,
		participationRepo: participationRepo,
		challengeRepo:     challengeRepo,
	}
}

func (s *leaderboardService) GetAbsoluteLeaderboard(ctx context.Context, challengeID, userID uuid.UUID, limit int64) (*models.AbsoluteLeaderboardResponse, error) {
	if challengeID == uuid.Nil || userID == uuid.Nil {
		return nil, fmt.Errorf("missing challenge or user id: %w", domain.ErrInvalidInput)
	}
	if limit <= 0 {
		limit = 20
	}

	top, err := s.leaderboardRepo.GetTopN(ctx, challengeID, limit)
	if err != nil {
		return nil, err
	}
	total, err := s.leaderboardRepo.GetTotalCount(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	var myPosition *models.LeaderboardEntry
	myRank, err := s.leaderboardRepo.GetRank(ctx, challengeID, userID)
	if err == nil {
		me, getErr := s.participationRepo.Get(ctx, userID, challengeID)
		if getErr == nil {
			myPosition = &models.LeaderboardEntry{
				UserID: userID,
				Score:  float64(me.CurrentScore),
				Rank:   myRank,
			}
		} else if !errors.Is(getErr, domain.ErrNotFound) {
			return nil, getErr
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	return &models.AbsoluteLeaderboardResponse{
		ChallengeID: challengeID,
		Type:        "absolute",
		Top:         top,
		MyPosition:  myPosition,
		TotalCount:  total,
	}, nil
}

func (s *leaderboardService) GetRelativeLeaderboard(ctx context.Context, challengeID, userID uuid.UUID, radius int64) (*models.RelativeLeaderboardResponse, error) {
	if challengeID == uuid.Nil || userID == uuid.Nil {
		return nil, fmt.Errorf("missing challenge or user id: %w", domain.ErrInvalidInput)
	}
	if radius <= 0 {
		radius = 2
	}

	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	meParticipation, err := s.participationRepo.Get(ctx, userID, challengeID)
	if err != nil {
		return nil, err
	}
	meRank, err := s.leaderboardRepo.GetRank(ctx, challengeID, userID)
	if err != nil {
		return nil, err
	}

	creatorScore := 0.0
	creatorRank := 0
	if creatorParticipation, getErr := s.participationRepo.Get(ctx, challenge.CreatorID, challengeID); getErr == nil {
		creatorScore = float64(creatorParticipation.CurrentScore)
		if rank, rankErr := s.leaderboardRepo.GetRank(ctx, challengeID, challenge.CreatorID); rankErr == nil {
			creatorRank = rank
		}
	} else if !errors.Is(getErr, domain.ErrNotFound) {
		return nil, getErr
	}

	nearby, err := s.leaderboardRepo.GetAroundUser(ctx, challengeID, userID, radius)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if nearby == nil {
		nearby = []models.LeaderboardEntry{}
	}

	resp := &models.RelativeLeaderboardResponse{
		ChallengeID: challengeID,
		Creator: models.LeaderboardEntry{
			UserID: challenge.CreatorID,
			Score:  creatorScore,
			Rank:   creatorRank,
		},
		Me: models.LeaderboardEntry{
			UserID: userID,
			Score:  float64(meParticipation.CurrentScore),
			Rank:   meRank,
		},
		Nearby: nearby,
	}

	if creatorScore > 0 {
		resp.RelativeToCreator.Percentage = round2((float64(meParticipation.CurrentScore) / creatorScore) * 100)
	}

	return resp, nil
}
