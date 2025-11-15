package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserVerification struct {
	ID          uuid.UUID  `db:"id"`
	UserID      uuid.UUID  `db:"user_id"`
	Email       string     `db:"email"`
	Code        string     `db:"code"`
	Attempts    int        `db:"attempts"`
	Confirmed   bool       `db:"confirmed"`
	ConfirmedAt *time.Time `db:"confirmed_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}
