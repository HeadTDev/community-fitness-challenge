package models

import (
	"time"

	"github.com/google/uuid"
)

type Participation struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	ChallengeID  uuid.UUID `json:"challenge_id" db:"challenge_id"`
	CurrentScore int       `json:"current_score" db:"current_score"`
	Rank         int       `json:"rank" db:"rank"`
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ParticipationResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	ChallengeID  uuid.UUID `json:"challenge_id"`
	CurrentScore int       `json:"current_score"`
	Rank         int       `json:"rank"`
	JoinedAt     time.Time `json:"joined_at"`
}

func (p *Participation) ToResponse() ParticipationResponse {
	return ParticipationResponse{
		UserID:       p.UserID,
		ChallengeID:  p.ChallengeID,
		CurrentScore: p.CurrentScore,
		Rank:         p.Rank,
		JoinedAt:     p.JoinedAt,
	}
}
