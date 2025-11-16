package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TargetGroup string

const (
	Pensioners    TargetGroup = "pensioners"
	Disabled      TargetGroup = "disabled"
	YoungFamilies TargetGroup = "young_families"
	LowIncome     TargetGroup = "low_income"
	Students      TargetGroup = "students"
	LargeFamilies TargetGroup = "large_families"
	Children      TargetGroup = "children"
	Veterans      TargetGroup = "veterans"
)

type BenefitLevel string

const (
	Regional   BenefitLevel = "regional"
	Federal    BenefitLevel = "federal"
	Commercial BenefitLevel = "commercial"
)

// TargetGroupList - кастомный тип для работы с JSON в БД
type TargetGroupList []TargetGroup

// Scan implements sql.Scanner interface
func (t *TargetGroupList) Scan(value interface{}) error {
	if value == nil {
		*t = []TargetGroup{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TargetGroupList: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, t)
}

// Value implements driver.Valuer interface
func (t TargetGroupList) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return json.Marshal(t)
}

// RegionList - кастомный тип для работы с JSON в БД (массив ID регионов)
type RegionList []int

// Scan implements sql.Scanner interface
func (r *RegionList) Scan(value interface{}) error {
	if value == nil {
		*r = []int{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan RegionList: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer interface
func (r RegionList) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

type BenefitTag string

const (
	MostPopular BenefitTag = "most_popular"
	New         BenefitTag = "new"
	Hot         BenefitTag = "hot"
	Best        BenefitTag = "best"
	Recommended BenefitTag = "recommended"
	Popular     BenefitTag = "popular"
	Top         BenefitTag = "top"
)

type BenefitTagList []BenefitTag

// Scan implements sql.Scanner interface
func (t *BenefitTagList) Scan(value interface{}) error {
	if value == nil {
		*t = []BenefitTag{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan BenefitTagList: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, t)
}

// Value implements driver.Valuer interface
func (t BenefitTagList) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return json.Marshal(t)
}

type Category string

const (
	Medicine  Category = "medicine"
	Transport Category = "transport"
	Food      Category = "food"
	Clothing  Category = "clothing"
	Other     Category = "other"
)

type Benefit struct {
	ID          uuid.UUID  `db:"id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	ValidFrom   time.Time  `db:"valid_from"`
	ValidTo     time.Time  `db:"valid_to"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"` // nullable

	Type           BenefitLevel    `db:"type"`
	TargetGroupIDs TargetGroupList `db:"target_group_ids"` // stored as JSON

	Longitude *float64 `db:"longitude"` // nullable
	Latitude  *float64 `db:"latitude"`  // nullable

	CityID *uuid.UUID `db:"city_id"` // nullable
	Region RegionList `db:"region"`  // stored as JSON array of region IDs

	Category    *Category      `db:"category"`   // nullable
	Requirement string         `db:"requirment"` // notice spelling to match table
	HowToUse    *string        `db:"how_to_use"` // nullable
	SourceURL   string         `db:"source_url"`
	Tags        BenefitTagList `db:"tags"` // stored as JSON array of tags

	Views int `db:"views"` // количество просмотров
}

type Favorite struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	BenefitID uuid.UUID  `db:"benefit_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"` // nullable
}

func (f *Favorite) IsDeleted() bool {
	return f.DeletedAt != nil
}
