package domain

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GroupType string

const (
	UserGroupPensioners    GroupType = "pensioners"
	UserGroupDisabled      GroupType = "disabled"
	UserGroupYoungFamilies GroupType = "young_families"
	UserGroupLowIncome     GroupType = "low_income"
	UserGroupStudents      GroupType = "students"
	UserGroupLargeFamilies GroupType = "large_families"
	UserGroupChildren      GroupType = "children"
	UserGroupVeterans      GroupType = "veterans"
)

type GroupTypeList []GroupType

// Value реализует интерфейс driver.Valuer для сохранения в БД
func (g GroupTypeList) Value() (driver.Value, error) {
	if g == nil {
		return nil, nil
	}
	return json.Marshal(g)
}

// Scan реализует интерфейс sql.Scanner для чтения из БД
func (g *GroupTypeList) Scan(value interface{}) error {
	if value == nil {
		*g = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for GroupTypeList: %T", value)
	}

	return json.Unmarshal(bytes, g)
}

type User struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	ExternalID  sql.NullString `db:"external_id" json:"external_id"`
	FirstName   sql.NullString `db:"first_name" json:"first_name"`
	LastName    sql.NullString `db:"last_name" json:"last_name"`
	MiddleName  sql.NullString `db:"middle_name" json:"middle_name"`
	SNILS       sql.NullString `db:"snils" json:"snils"`
	Email       sql.NullString `db:"email" json:"email"`
	PhoneNumber sql.NullString `db:"phone_number" json:"phone_number"`
	CityID      *uuid.UUID     `db:"city_id" json:"city_id"`
	GroupType   GroupTypeList  `db:"group_type" json:"group_type"`

	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	RegisteredAt *time.Time `db:"registered_at" json:"registered_at,omitempty"`
}
