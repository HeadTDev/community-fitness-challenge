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
	primaryRepo       repositories.LeaderboardRepository
	fallbackRepo      repositories.LeaderboardRepository
	participationRepo repositories.ParticipationRepository
	challengeRepo     repositories.ChallengeRepository
}

func NewLeaderboardService(
	primaryRepo repositories.LeaderboardRepository,
	fallbackRepo repositories.LeaderboardRepository,
	participationRepo repositories.ParticipationRepository,
	challengeRepo repositories.ChallengeRepository,
) LeaderboardService {
	return &leaderboardService{
		primaryRepo:       primaryRepo,
		fallbackRepo:      fallbackRepo,
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

	top, total, err := s.getTopWithCount(ctx, challengeID, limit)
	if err != nil {
		return nil, err
	}

	var myPosition *models.LeaderboardEntry
	myRank, err := s.getRankWithFallback(ctx, challengeID, userID)
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
	meRank, err := s.getRankWithFallback(ctx, challengeID, userID)
	if err != nil {
		return nil, err
	}

	creatorScore := 0.0
	creatorRank := 0
	if creatorParticipation, getErr := s.participationRepo.Get(ctx, challenge.CreatorID, challengeID); getErr == nil {
		creatorScore = float64(creatorParticipation.CurrentScore)
		if rank, rankErr := s.getRankWithFallback(ctx, challengeID, challenge.CreatorID); rankErr == nil {
			creatorRank = rank
		}
	} else if !errors.Is(getErr, domain.ErrNotFound) {
		return nil, getErr
	}

	nearby, err := s.getAroundWithFallback(ctx, challengeID, userID, radius)
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

func (s *leaderboardService) getTopWithCount(ctx context.Context, challengeID uuid.UUID, limit int64) ([]models.LeaderboardEntry, int64, error) {
	top, topErr := s.primaryRepo.GetTopN(ctx, challengeID, limit)
	total, totalErr := s.primaryRepo.GetTotalCount(ctx, challengeID)
	if topErr == nil && totalErr == nil && len(top) > 0 {
		return top, total, nil
	}

	fallbackTop, err := s.fallbackRepo.GetTopN(ctx, challengeID, limit)
	if err != nil {
		if topErr != nil {
			return nil, 0, topErr
		}
		return nil, 0, err
	}
	fallbackTotal, err := s.fallbackRepo.GetTotalCount(ctx, challengeID)
	if err != nil {
		return nil, 0, err
	}
	return fallbackTop, fallbackTotal, nil
}

func (s *leaderboardService) getRankWithFallback(ctx context.Context, challengeID, userID uuid.UUID) (int, error) {
	rank, err := s.primaryRepo.GetRank(ctx, challengeID, userID)
	if err == nil {
		return rank, nil
	}
	fallbackRank, fbErr := s.fallbackRepo.GetRank(ctx, challengeID, userID)
	if fbErr == nil {
		return fallbackRank, nil
	}
	return 0, err
}

func (s *leaderboardService) getAroundWithFallback(ctx context.Context, challengeID, userID uuid.UUID, radius int64) ([]models.LeaderboardEntry, error) {
	nearby, err := s.primaryRepo.GetAroundUser(ctx, challengeID, userID, radius)
	if err == nil && len(nearby) > 0 {
		return nearby, nil
	}
	fallbackNearby, fbErr := s.fallbackRepo.GetAroundUser(ctx, challengeID, userID, radius)
	if fbErr == nil {
		return fallbackNearby, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fbErr
}
