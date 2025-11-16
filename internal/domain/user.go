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

// Статус подтверждения группы
type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"  // Ожидает подтверждения
	VerificationStatusVerified VerificationStatus = "verified" // Подтверждена
	VerificationStatusRejected VerificationStatus = "rejected" // Отклонена
	VerificationStatusExpired  VerificationStatus = "expired"  // Истекла
)

// Группа пользователя со статусом подтверждения
type UserGroup struct {
	Type         GroupType          `json:"type"`
	Status       VerificationStatus `json:"status"`
	VerifiedAt   *time.Time         `json:"verified_at,omitempty"`   // Когда подтверждена
	RejectedAt   *time.Time         `json:"rejected_at,omitempty"`   // Когда отклонена
	ExpiresAt    *time.Time         `json:"expires_at,omitempty"`    // Когда истекает
	ExternalID   string             `json:"external_id,omitempty"`   // ID из внешнего API
	ErrorMessage string             `json:"error_message,omitempty"` // Сообщение об ошибке
}

// Список групп со статусами
type UserGroupList []UserGroup

// Value реализует интерфейс driver.Valuer для сохранения в БД
func (g UserGroupList) Value() (driver.Value, error) {
	if g == nil {
		return nil, nil
	}
	return json.Marshal(g)
}

// Scan реализует интерфейс sql.Scanner для чтения из БД
func (g *UserGroupList) Scan(value interface{}) error {
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
		return fmt.Errorf("unsupported type for UserGroupList: %T", value)
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
	GroupType   UserGroupList  `db:"group_type" json:"group_type"`

	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	RegisteredAt *time.Time `db:"registered_at" json:"registered_at,omitempty"`
	Documents    []UserDocument
}

type UserDocumentType string

const (
	UserDocumentTypePassport     UserDocumentType = "passport"
	UserDocumentTypeSNILS        UserDocumentType = "snils"
	UserDocumentTypeRegistration UserDocumentType = "registration"
)

type UserDocument struct {
	ID             uuid.UUID        `db:"id" json:"id"`
	UserID         uuid.UUID        `db:"user_id" json:"user_id"`
	DocumentType   UserDocumentType `db:"document_type" json:"document_type"`
	DocumentNumber string           `db:"document_number" json:"document_number"`
	CreatedAt      time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time       `db:"deleted_at" json:"deleted_at,omitempty"`
}
