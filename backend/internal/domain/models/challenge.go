package models

import (
	"time"

	"github.com/google/uuid"
)

type ChallengeStatus string

const (
	ChallengeStatusUpcoming ChallengeStatus = "upcoming"
	ChallengeStatusActive   ChallengeStatus = "active"
	ChallengeStatusFinished ChallengeStatus = "finished"
)

type ChallengeType string

const (
	ChallengeTypeSteps         ChallengeType = "steps"
	ChallengeTypeCalories      ChallengeType = "calories"
	ChallengeTypeActiveMinutes ChallengeType = "active_minutes"
	ChallengeTypeMixed         ChallengeType = "mixed"
)

type Challenge struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Title       string          `json:"title" db:"title"`
	Description *string         `json:"description,omitempty" db:"description"`
	ImageURL    *string         `json:"image_url,omitempty" db:"image_url"`
	StartDate   time.Time       `json:"start_date" db:"start_date"`
	EndDate     time.Time       `json:"end_date" db:"end_date"`
	Status      ChallengeStatus `json:"status" db:"status"`
	Type        ChallengeType   `json:"type" db:"type"`
	Goal        int             `json:"goal" db:"goal"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ChallengeResponse is a DTO for returning challenge data via API.
type ChallengeResponse struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	ImageURL    *string         `json:"image_url,omitempty"`
	StartDate   time.Time       `json:"start_date"`
	EndDate     time.Time       `json:"end_date"`
	Status      ChallengeStatus `json:"status"`
	Type        ChallengeType   `json:"type"`
	Goal        int             `json:"goal"`
}

func (c *Challenge) ToResponse() ChallengeResponse {
	return ChallengeResponse{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		ImageURL:    c.ImageURL,
		StartDate:   c.StartDate,
		EndDate:     c.EndDate,
		Status:      c.Status,
		Type:        c.Type,
		Goal:        c.Goal,
	}
}
