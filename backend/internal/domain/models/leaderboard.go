package models

import "github.com/google/uuid"

type LeaderboardEntry struct {
	UserID uuid.UUID `json:"user_id"`
	Score  float64   `json:"score"`
	Rank   int       `json:"rank"`
}
