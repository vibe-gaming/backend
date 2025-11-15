package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID      `db:"id"`
	ExternalID  sql.NullString `db:"external_id"`
	FirstName   sql.NullString `db:"first_name"`
	LastName    sql.NullString `db:"last_name"`
	MiddleName  sql.NullString `db:"middle_name"`
	SNILS       sql.NullString `db:"snils"`
	Email       sql.NullString `db:"email"`
	PhoneNumber sql.NullString `db:"phone_number"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}
