package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRegistration struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	Phone       string     `json:"phone" db:"phone"`
	Password    string     `json:"password" db:"password"`
	Code        string     `json:"code" db:"code"`
	Confirmed   bool       `json:"confirmed" db:"confirmed"`
	ConfirmedAt *time.Time `json:"confirmed_at" db:"confirmed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
}
