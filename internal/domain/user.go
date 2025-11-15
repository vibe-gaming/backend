package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	Password string    `db:"password"`
	Email    string    `db:"email"`

	// ESIA OAuth fields
	ESIAOID        sql.NullString `db:"esia_oid"`
	ESIAFirstName  sql.NullString `db:"esia_first_name"`
	ESIALastName   sql.NullString `db:"esia_last_name"`
	ESIAMiddleName sql.NullString `db:"esia_middle_name"`
	ESIASNILS      sql.NullString `db:"esia_snils"`
	ESIAEmail      sql.NullString `db:"esia_email"`
	ESIAMobile     sql.NullString `db:"esia_mobile"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}
