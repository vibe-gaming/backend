package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	RefreshToken uuid.UUID  `json:"refresh_token" db:"refresh_token"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	IP           string     `json:"ip" db:"ip"`
	ExpiresIn    time.Time  `json:"expires_in" db:"expires_in"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`
}
