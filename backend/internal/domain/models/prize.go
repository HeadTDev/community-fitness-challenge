package models

import (
	"time"

	"github.com/google/uuid"
)

type Prize struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ChallengeID  uuid.UUID `json:"challenge_id" db:"challenge_id"`
	Title        string    `json:"title" db:"title"`
	Description  *string   `json:"description,omitempty" db:"description"`
	ImageURL     *string   `json:"image_url,omitempty" db:"image_url"`
	RankRequired int       `json:"rank_required" db:"rank_required"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type PrizeResponse struct {
	ID           uuid.UUID `json:"id"`
	ChallengeID  uuid.UUID `json:"challenge_id"`
	Title        string    `json:"title"`
	Description  *string   `json:"description,omitempty"`
	ImageURL     *string   `json:"image_url,omitempty"`
	RankRequired int       `json:"rank_required"`
}

func (p *Prize) ToResponse() PrizeResponse {
	return PrizeResponse{
		ID:           p.ID,
		ChallengeID:  p.ChallengeID,
		Title:        p.Title,
		Description:  p.Description,
		ImageURL:     p.ImageURL,
		RankRequired: p.RankRequired,
	}
}
