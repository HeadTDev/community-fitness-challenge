package models

import "github.com/google/uuid"

type LeaderboardEntry struct {
	UserID uuid.UUID `json:"user_id"`
	Score  float64   `json:"score"`
	Rank   int       `json:"rank"`
}

type AbsoluteLeaderboardResponse struct {
	ChallengeID uuid.UUID          `json:"challenge_id"`
	Type        string             `json:"type"`
	Top         []LeaderboardEntry `json:"top"`
	MyPosition  *LeaderboardEntry  `json:"my_position,omitempty"`
	TotalCount  int64              `json:"total_count"`
}

type RelativeLeaderboardResponse struct {
	ChallengeID       uuid.UUID          `json:"challenge_id"`
	Creator           LeaderboardEntry   `json:"creator"`
	Me                LeaderboardEntry   `json:"me"`
	Nearby            []LeaderboardEntry `json:"nearby"`
	RelativeToCreator struct {
		Percentage float64 `json:"percentage"`
	} `json:"relative_to_creator"`
}
