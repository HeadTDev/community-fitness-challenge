package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleParticipant UserRole = "participant"
	RoleCreator     UserRole = "creator"
	RoleAdmin       UserRole = "admin"
)

type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	AppleID     *string    `json:"apple_id,omitempty" db:"apple_id"`
	Email       string     `json:"email" db:"email"`
	DisplayName *string    `json:"display_name,omitempty" db:"display_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio         *string    `json:"bio,omitempty" db:"bio"`
	Timezone    string     `json:"timezone" db:"timezone"`
	Role        UserRole   `json:"role" db:"role"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// S3 Constants for user assets
const (
	AvatarBucket = "fitchallenge-assets"
	AvatarPrefix = "avatars"
)

// UserResponse is a DTO for returning user data via API.
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	Bio         *string   `json:"bio"`
	Timezone    string    `json:"timezone"`
	Role        UserRole  `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Bio:         u.Bio,
		Timezone:    u.Timezone,
		Role:        u.Role,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
