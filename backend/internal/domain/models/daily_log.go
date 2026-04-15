package models

import (
	"time"

	"github.com/google/uuid"
)

type DailyLog struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	ChallengeID       uuid.UUID `json:"challenge_id" db:"challenge_id"`
	LogDate           time.Time `json:"log_date" db:"log_date"`
	Steps             int       `json:"steps" db:"steps"`
	Calories          int       `json:"calories" db:"calories"`
	ActiveMinutes     int       `json:"active_minutes" db:"active_minutes"`
	Score             float64   `json:"score" db:"score"`
	HealthKitDataHash *string   `json:"healthkit_data_hash,omitempty" db:"healthkit_data_hash"`
	SourceBundleIDs   []string  `json:"source_bundle_ids,omitempty" db:"source_bundle_ids"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type DailyLogResponse struct {
	ID            uuid.UUID `json:"id"`
	LogDate       time.Time `json:"log_date"`
	Steps         int       `json:"steps"`
	Calories      int       `json:"calories"`
	ActiveMinutes int       `json:"active_minutes"`
	Score         float64   `json:"score"`
}

type DailyLogAggregation struct {
	TotalScore    float64 `json:"total_score"`
	TotalCalories int     `json:"total_calories"`
	DaysLogged    int     `json:"days_logged"`
	Streak        int     `json:"streak"`
}

type DailyLogListResponse struct {
	Logs        []DailyLogResponse  `json:"logs"`
	Aggregation DailyLogAggregation `json:"aggregation"`
}

func (d *DailyLog) ToResponse() DailyLogResponse {
	return DailyLogResponse{
		ID:            d.ID,
		LogDate:       d.LogDate,
		Steps:         d.Steps,
		Calories:      d.Calories,
		ActiveMinutes: d.ActiveMinutes,
		Score:         d.Score,
	}
}
