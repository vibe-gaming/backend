package domain

import (
	"time"

	"github.com/google/uuid"
)

type City struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	RegionID  uuid.UUID  `db:"region_id" json:"region_id"`
	Name      string     `db:"name" json:"name"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
