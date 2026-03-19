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
	Timezone    string     `json:"timezone" db:"timezone"`
	Role        UserRole   `json:"role" db:"role"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
