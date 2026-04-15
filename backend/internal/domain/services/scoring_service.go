package services

import (
	"fmt"
	"math"

	"github.com/HeadTDev/fitchallenge/internal/domain"
)

const (
	ScoringVersionV1 = "v1"

	stepsTargetV1         = 15000.0
	caloriesTargetV1      = 1000.0
	activeMinutesTargetV1 = 60.0

	stepsWeightV1         = 0.30
	caloriesWeightV1      = 0.40
	activeMinutesWeightV1 = 0.30
)

type ScoreInput struct {
	Steps         int
	Calories      int
	ActiveMinutes int
}

type ScoreBreakdown struct {
	Version string  `json:"version"`
	Score   float64 `json:"score"`

	StepsNormalized         float64 `json:"steps_normalized"`
	CaloriesNormalized      float64 `json:"calories_normalized"`
	ActiveMinutesNormalized float64 `json:"active_minutes_normalized"`

	StepsContribution         float64 `json:"steps_contribution"`
	CaloriesContribution      float64 `json:"calories_contribution"`
	ActiveMinutesContribution float64 `json:"active_minutes_contribution"`
}

type ScoringService interface {
	Calculate(input ScoreInput) (*ScoreBreakdown, error)
}

type scoringService struct{}

func NewScoringService() ScoringService {
	return &scoringService{}
}

func (s *scoringService) Calculate(input ScoreInput) (*ScoreBreakdown, error) {
	if input.Steps < 0 || input.Calories < 0 || input.ActiveMinutes < 0 {
		return nil, fmt.Errorf("negative metrics are not allowed: %w", domain.ErrInvalidInput)
	}
	if input.Steps == 0 && input.Calories == 0 && input.ActiveMinutes == 0 {
		return nil, fmt.Errorf("at least one metric must be greater than zero: %w", domain.ErrInvalidInput)
	}

	stepsNorm := normalize(float64(input.Steps), stepsTargetV1)
	caloriesNorm := normalize(float64(input.Calories), caloriesTargetV1)
	activeNorm := normalize(float64(input.ActiveMinutes), activeMinutesTargetV1)

	stepsContribution := stepsNorm * stepsWeightV1
	caloriesContribution := caloriesNorm * caloriesWeightV1
	activeContribution := activeNorm * activeMinutesWeightV1

	score := round2((stepsContribution + caloriesContribution + activeContribution) * 100)

	return &ScoreBreakdown{
		Version: ScoringVersionV1,
		Score:   score,

		StepsNormalized:         round4(stepsNorm),
		CaloriesNormalized:      round4(caloriesNorm),
		ActiveMinutesNormalized: round4(activeNorm),

		StepsContribution:         round4(stepsContribution),
		CaloriesContribution:      round4(caloriesContribution),
		ActiveMinutesContribution: round4(activeContribution),
	}, nil
}

func normalize(value, target float64) float64 {
	if target <= 0 {
		return 0
	}
	if value <= 0 {
		return 0
	}
	norm := value / target
	if norm > 1 {
		return 1
	}
	return norm
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
