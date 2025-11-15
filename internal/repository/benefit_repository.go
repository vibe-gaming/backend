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
	Type         *string
	TargetGroups []string
	DateFrom     *string
	DateTo       *string
	Search       *string
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
	INSERT INTO benefit (id, title, description, valid_from, valid_to, created_at, updated_at, deleted_at, type, target_group_ids, longitude, latitude, region, requirment, how_to_use, source_url)
	VALUES (uuid_to_bin(?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, benefit.ID, benefit.Title, benefit.Description, benefit.ValidFrom, benefit.ValidTo, benefit.CreatedAt, benefit.UpdatedAt, benefit.DeletedAt, benefit.Type, benefit.TargetGroupIDs, benefit.Longitude, benefit.Latitude, benefit.Region, benefit.Requirement, benefit.HowToUse, benefit.SourceURL)
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
			region,
			requirment,
			how_to_use,
			source_url
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
			region,
			requirment,
			how_to_use,
			source_url
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

		// Фильтр по датам (льгота должна быть активна в указанном периоде)
		if filters.DateFrom != nil {
			query += ` AND valid_to >= ?`
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += ` AND valid_from <= ?`
			args = append(args, *filters.DateTo)
		}

		// Текстовый поиск по названию и описанию
		if filters.Search != nil && *filters.Search != "" {
			searchTerm := "%" + *filters.Search + "%"
			query += ` AND (title LIKE ? OR description LIKE ?)`
			args = append(args, searchTerm, searchTerm)
		}
	}

	query += `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

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

		// Фильтр по датам
		if filters.DateFrom != nil {
			query += ` AND valid_to >= ?`
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += ` AND valid_from <= ?`
			args = append(args, *filters.DateTo)
		}

		// Текстовый поиск
		if filters.Search != nil && *filters.Search != "" {
			searchTerm := "%" + *filters.Search + "%"
			query += ` AND (title LIKE ? OR description LIKE ?)`
			args = append(args, searchTerm, searchTerm)
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
			region = ?,
			requirment = ?,
			how_to_use = ?,
			source_url = ?
		WHERE id = uuid_to_bin(?)
	`
	_, err := r.db.ExecContext(ctx, query, benefit.Title, benefit.Description, benefit.ValidFrom, benefit.ValidTo, benefit.UpdatedAt, benefit.DeletedAt, benefit.Type, benefit.TargetGroupIDs, benefit.Longitude, benefit.Latitude, benefit.Region, benefit.Requirement, benefit.HowToUse, benefit.SourceURL, benefit.ID)
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
