package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
)

type BenefitFilters struct {
	RegionID     *int
	CityID       *string
	Type         *string
	TargetGroups []string
	Tags         []string
	Categories   []string
	DateFrom     *string
	DateTo       *string
	Search       *string
	SearchMode   string // "natural" или "boolean"
	SortBy       string // "created_at", "views", "updated_at"
	Order        string // "asc", "desc"
}

type BenefitRepository interface {
	Create(ctx context.Context, benefit *domain.Benefit) error
	GetByID(ctx context.Context, id string) (*domain.Benefit, error)
	GetAll(ctx context.Context, limit, offset int, filters *BenefitFilters) ([]*domain.Benefit, error)
	Count(ctx context.Context, filters *BenefitFilters) (int64, error)
	Update(ctx context.Context, benefit *domain.Benefit) error
	Delete(ctx context.Context, id string) error
}

type benefitRepository struct {
	db *sqlx.DB
}

func NewBenefitRepository(db *sqlx.DB) BenefitRepository {
	return &benefitRepository{
		db: db,
	}
}

func (r *benefitRepository) Create(ctx context.Context, benefit *domain.Benefit) error {
	const query = `
	INSERT INTO benefit (id, title, description, valid_from, valid_to, created_at, updated_at, deleted_at, type, target_group_ids, longitude, latitude, city_id, region, category, requirment, how_to_use, source_url, tags, views)
	VALUES (uuid_to_bin(?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, uuid_to_bin(?), ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, benefit.ID, benefit.Title, benefit.Description, benefit.ValidFrom, benefit.ValidTo, benefit.CreatedAt, benefit.UpdatedAt, benefit.DeletedAt, benefit.Type, benefit.TargetGroupIDs, benefit.Longitude, benefit.Latitude, benefit.CityID, benefit.Region, benefit.Category, benefit.Requirement, benefit.HowToUse, benefit.SourceURL, benefit.Tags, benefit.Views)
	if err != nil {
		return fmt.Errorf("db insert benefit: %w", err)
	}
	return nil
}
func (r *benefitRepository) GetByID(ctx context.Context, id string) (*domain.Benefit, error) {
	const query = `
		SELECT 
			bin_to_uuid(id) as id,
			title,
			description,
			valid_from,
			valid_to,
			created_at,
			updated_at,
			deleted_at,
			type,
			target_group_ids,
			longitude,
			latitude,
			bin_to_uuid(city_id) as city_id,
			region,
			category,
			requirment,
			how_to_use,
			source_url,
			tags,
			views
		FROM benefit
		WHERE id = uuid_to_bin(?) AND deleted_at IS NULL
	`
	var benefit domain.Benefit
	err := r.db.GetContext(ctx, &benefit, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &benefit, nil
}

func (r *benefitRepository) GetAll(ctx context.Context, limit, offset int, filters *BenefitFilters) ([]*domain.Benefit, error) {
	query := `
		SELECT 
			bin_to_uuid(id) as id,
			title,
			description,
			valid_from,
			valid_to,
			created_at,
			updated_at,
			deleted_at,
			type,
			target_group_ids,
			longitude,
			latitude,
			bin_to_uuid(city_id) as city_id,
			region,
			category,
			requirment,
			how_to_use,
			source_url,
			tags,
			views
		FROM benefit
		WHERE deleted_at IS NULL`

	args := []interface{}{}

	// Применяем фильтры
	if filters != nil {
		// Фильтр по региону
		if filters.RegionID != nil {
			query += ` AND JSON_CONTAINS(region, ?)`
			args = append(args, fmt.Sprintf("%d", *filters.RegionID))
		}

		// Фильтр по городу
		if filters.CityID != nil {
			query += ` AND city_id = UUID_TO_BIN(?)`
			args = append(args, *filters.CityID)
		}

		// Фильтр по типу
		if filters.Type != nil {
			query += ` AND type = ?`
			args = append(args, *filters.Type)
		}

		// Фильтр по целевым группам (хотя бы одна группа должна совпадать)
		if len(filters.TargetGroups) > 0 {
			query += ` AND (`
			for i, group := range filters.TargetGroups {
				if i > 0 {
					query += ` OR `
				}
				query += `JSON_CONTAINS(target_group_ids, ?)`
				args = append(args, fmt.Sprintf(`"%s"`, group))
			}
			query += `)`
		}

		// Фильтр по тегам (хотя бы один тег должен совпадать)
		if len(filters.Tags) > 0 {
			query += ` AND (`
			for i, tag := range filters.Tags {
				if i > 0 {
					query += ` OR `
				}
				query += `JSON_CONTAINS(tags, ?)`
				args = append(args, fmt.Sprintf(`"%s"`, tag))
			}
			query += `)`
		}

		// Фильтр по категориям (хотя бы одна категория должна совпадать)
		if len(filters.Categories) > 0 {
			query += ` AND (`
			for i, category := range filters.Categories {
				if i > 0 {
					query += ` OR `
				}
				query += `category = ?`
				args = append(args, category)
			}
			query += `)`
		}

		// Фильтр по датам (льгота должна быть активна в указанном периоде)
		if filters.DateFrom != nil {
			query += ` AND valid_to >= ?`
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += ` AND valid_from <= ?`
			args = append(args, *filters.DateTo)
		}

		// Текстовый поиск с Full-Text Search
		if filters.Search != nil && *filters.Search != "" {
			if filters.SearchMode == "boolean" {
				query += ` AND MATCH(title, description) AGAINST(? IN BOOLEAN MODE)`
			} else {
				// По умолчанию используем NATURAL LANGUAGE MODE
				query += ` AND MATCH(title, description) AGAINST(? IN NATURAL LANGUAGE MODE)`
			}
			args = append(args, *filters.Search)
		}
	}

	// Сортировка
	orderBy := "created_at"
	orderDir := "DESC"

	if filters != nil {
		// Определяем поле для сортировки
		switch filters.SortBy {
		case "views":
			orderBy = "views"
		case "updated_at":
			orderBy = "updated_at"
		case "created_at":
			orderBy = "created_at"
		default:
			orderBy = "created_at"
		}

		// Определяем направление сортировки
		if filters.Order == "asc" {
			orderDir = "ASC"
		} else {
			orderDir = "DESC"
		}
	}

	query += fmt.Sprintf(`
		ORDER BY %s %s
		LIMIT ? OFFSET ?`, orderBy, orderDir)

	args = append(args, limit, offset)

	var benefits []*domain.Benefit
	err := r.db.SelectContext(ctx, &benefits, query, args...)
	if err != nil {
		return nil, err
	}
	return benefits, nil
}

func (r *benefitRepository) Count(ctx context.Context, filters *BenefitFilters) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM benefit 
		WHERE deleted_at IS NULL`

	args := []interface{}{}

	// Применяем те же фильтры
	if filters != nil {
		// Фильтр по региону
		if filters.RegionID != nil {
			query += ` AND JSON_CONTAINS(region, ?)`
			args = append(args, fmt.Sprintf("%d", *filters.RegionID))
		}

		// Фильтр по городу
		if filters.CityID != nil {
			query += ` AND city_id = UUID_TO_BIN(?)`
			args = append(args, *filters.CityID)
		}

		// Фильтр по типу
		if filters.Type != nil {
			query += ` AND type = ?`
			args = append(args, *filters.Type)
		}

		// Фильтр по целевым группам
		if len(filters.TargetGroups) > 0 {
			query += ` AND (`
			for i, group := range filters.TargetGroups {
				if i > 0 {
					query += ` OR `
				}
				query += `JSON_CONTAINS(target_group_ids, ?)`
				args = append(args, fmt.Sprintf(`"%s"`, group))
			}
			query += `)`
		}

		// Фильтр по тегам
		if len(filters.Tags) > 0 {
			query += ` AND (`
			for i, tag := range filters.Tags {
				if i > 0 {
					query += ` OR `
				}
				query += `JSON_CONTAINS(tags, ?)`
				args = append(args, fmt.Sprintf(`"%s"`, tag))
			}
			query += `)`
		}

		// Фильтр по категориям
		if len(filters.Categories) > 0 {
			query += ` AND (`
			for i, category := range filters.Categories {
				if i > 0 {
					query += ` OR `
				}
				query += `category = ?`
				args = append(args, category)
			}
			query += `)`
		}

		// Фильтр по датам
		if filters.DateFrom != nil {
			query += ` AND valid_to >= ?`
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += ` AND valid_from <= ?`
			args = append(args, *filters.DateTo)
		}

		// Текстовый поиск с Full-Text Search
		if filters.Search != nil && *filters.Search != "" {
			if filters.SearchMode == "boolean" {
				query += ` AND MATCH(title, description) AGAINST(? IN BOOLEAN MODE)`
			} else {
				query += ` AND MATCH(title, description) AGAINST(? IN NATURAL LANGUAGE MODE)`
			}
			args = append(args, *filters.Search)
		}
	}

	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *benefitRepository) Update(ctx context.Context, benefit *domain.Benefit) error {
	const query = `
		UPDATE benefit
		SET
			title = ?,
			description = ?,
			valid_from = ?,
			valid_to = ?,
			updated_at = ?,
			deleted_at = ?,
			type = ?,
			target_group_ids = ?,
			longitude = ?,
			latitude = ?,
			city_id = uuid_to_bin(?),
			region = ?,
			category = ?,
			requirment = ?,
			how_to_use = ?,
			source_url = ?,
			tags = ?,
			views = ?
		WHERE id = uuid_to_bin(?)
	`
	_, err := r.db.ExecContext(ctx, query, benefit.Title, benefit.Description, benefit.ValidFrom, benefit.ValidTo, benefit.UpdatedAt, benefit.DeletedAt, benefit.Type, benefit.TargetGroupIDs, benefit.Longitude, benefit.Latitude, benefit.CityID, benefit.Region, benefit.Category, benefit.Requirement, benefit.HowToUse, benefit.SourceURL, benefit.Tags, benefit.Views, benefit.ID)
	if err != nil {
		return fmt.Errorf("db update benefit: %w", err)
	}
	return nil
}

func (r *benefitRepository) Delete(ctx context.Context, id string) error {
	const query = `
		UPDATE benefit
		SET deleted_at = NOW()
		WHERE id = uuid_to_bin(?) AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("db delete benefit: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("db delete benefit: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
