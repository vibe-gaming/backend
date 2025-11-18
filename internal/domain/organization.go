package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`

	Buildings []OrganizationBuilding `db:"buildings"`
}

type OrganizationType string

const (
	OrganizationTypeFarmacy = "farmacy"
)

type OrganizationTag string

const (
	OrganizationTagIsHaveRamp = "is_have_ramp"
	OrganizationTagIsHaveLift = "is_have_lift"
)

type OrganizationTagList []OrganizationTag

// Scan implements sql.Scanner interface
func (t *OrganizationTagList) Scan(value interface{}) error {
	if value == nil {
		*t = []OrganizationTag{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan OrganizationTagList: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, t)
}

// Value implements driver.Valuer interface
func (t OrganizationTagList) Value() (driver.Value, error) {
	if t == nil {
		return json.Marshal([]OrganizationTag{})
	}
	return json.Marshal(t)
}

type OrganizationBuilding struct {
	ID             uuid.UUID  `db:"id"`
	OrganizationID uuid.UUID  `db:"organization_id"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
	DeletedAt      *time.Time `db:"deleted_at"`

	Address     string    `db:"address"`
	Latitude    float64   `db:"latitude"`
	Longitude   float64   `db:"longitude"`
	PhoneNumber string    `db:"phone_number"`
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`
	IsOpen      bool      `db:"is_open"`
	Type        string    `db:"type"`

	Tags OrganizationTagList `db:"tags"`
}

func (o *OrganizationBuilding) GetGisDeeplink() string {

	if o.Longitude == 0 || o.Latitude == 0 {
		return ""
	}
	return fmt.Sprintf("https://2gis.ru/yakutsk/geo/%v,%v?m=%v,%v/17.38", o.Longitude, o.Latitude, o.Longitude, o.Latitude)
}

func (o *OrganizationBuilding) IsOpenAlright() bool {

	return true
}
