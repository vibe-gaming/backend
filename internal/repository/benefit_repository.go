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
	SearchMode   string  // "natural" или "boolean"
	SortBy       string  // "created_at", "views", "updated_at"
	Order        string  // "asc", "desc"
	UserID       *string // UUID пользователя для фильтрации избранных (favorites=true)
}

type BenefitRepository interface {
	Create(ctx context.Context, benefit *domain.Benefit) error
	GetByID(ctx context.Context, id string) (*domain.Benefit, error)
	GetAll(ctx context.Context, limit, offset int, filters *BenefitFilters) ([]*domain.Benefit, error)
	Count(ctx context.Context, filters *BenefitFilters) (int64, error)
	Update(ctx context.Context, benefit *domain.Benefit) error
	Delete(ctx context.Context, id string) error
	GetFilterStats(ctx context.Context, filters *BenefitFilters) (*FilterStats, error)
}

type FilterStats struct {
	Categories map[string]int64 `json:"categories"`
	Levels     map[string]int64 `json:"levels"`
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
			bin_to_uuid(b.id) as id,
			b.title,
			b.description,
			b.valid_from,
			b.valid_to,
			b.created_at,
			b.updated_at,
			b.deleted_at,
			b.type,
			b.target_group_ids,
			b.longitude,
			b.latitude,
			bin_to_uuid(b.city_id) as city_id,
			b.region,
			b.category,
			b.requirment,
			b.how_to_use,
			b.source_url,
			b.tags,
			b.views
		FROM benefit b`

	args := []interface{}{}

	// JOIN с таблицей favorite если нужно фильтровать по избранным
	if filters != nil && filters.UserID != nil {
		query += `
		INNER JOIN favorite f ON b.id = f.benefit_id 
			AND f.user_id = UUID_TO_BIN(?) 
			AND f.deleted_at IS NULL`
		args = append(args, *filters.UserID)
		fmt.Printf("DEBUG GetAll: Applying favorites filter for user_id=%s\n", *filters.UserID)
	}

	query += `
		WHERE b.deleted_at IS NULL`

	// Применяем фильтры
	if filters != nil {
		// Фильтр по региону
		if filters.RegionID != nil {
			query += ` AND JSON_CONTAINS(b.region, ?)`
			args = append(args, fmt.Sprintf("%d", *filters.RegionID))
		}

		// Фильтр по городу
		if filters.CityID != nil {
			query += ` AND b.city_id = UUID_TO_BIN(?)`
			args = append(args, *filters.CityID)
		}

		// Фильтр по типу
		if filters.Type != nil {
			query += ` AND b.type = ?`
			args = append(args, *filters.Type)
		}

		// Фильтр по целевым группам (хотя бы одна группа должна совпадать)
		if len(filters.TargetGroups) > 0 {
			query += ` AND (`
			for i, group := range filters.TargetGroups {
				if i > 0 {
					query += ` OR `
				}
				query += `JSON_CONTAINS(b.target_group_ids, ?)`
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
				query += `JSON_CONTAINS(b.tags, ?)`
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
				query += `b.category = ?`
				args = append(args, category)
			}
			query += `)`
		}

		// Фильтр по датам (льгота должна быть активна в указанном периоде)
		if filters.DateFrom != nil {
			query += ` AND b.valid_to >= ?`
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += ` AND b.valid_from <= ?`
			args = append(args, *filters.DateTo)
		}

		// Текстовый поиск с Full-Text Search
		if filters.Search != nil && *filters.Search != "" {
			if filters.SearchMode == "boolean" {
				query += ` AND MATCH(b.title, b.description) AGAINST(? IN BOOLEAN MODE)`
			} else {
				// По умолчанию используем NATURAL LANGUAGE MODE
				query += ` AND MATCH(b.title, b.description) AGAINST(? IN NATURAL LANGUAGE MODE)`
			}
			args = append(args, *filters.Search)
		}
	}

	// Сортировка
	orderBy := "b.created_at"
	orderDir := "DESC"

	if filters != nil {
		// Определяем поле для сортировки
		switch filters.SortBy {
		case "views":
			orderBy = "b.views"
		case "updated_at":
			orderBy = "b.updated_at"
		case "created_at":
			orderBy = "b.created_at"
		default:
			orderBy = "b.created_at"
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

	fmt.Printf("DEBUG GetAll: Final query:\n%s\n", query)
	fmt.Printf("DEBUG GetAll: Args: %+v\n", args)

	var benefits []*domain.Benefit
	err := r.db.SelectContext(ctx, &benefits, query, args...)
	if err != nil {
		fmt.Printf("DEBUG GetAll: Query error: %v\n", err)
		return nil, err
	}
	fmt.Printf("DEBUG GetAll: Found %d benefits\n", len(benefits))
	return benefits, nil
}

func (r *benefitRepository) Count(ctx context.Context, filters *BenefitFilters) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM benefit b`

	args := []interface{}{}

	// JOIN с таблицей favorite если нужно фильтровать по избранным
	if filters != nil && filters.UserID != nil {
		query += `
		INNER JOIN favorite f ON b.id = f.benefit_id 
			AND f.user_id = UUID_TO_BIN(?) 
			AND f.deleted_at IS NULL`
		args = append(args, *filters.UserID)
	}

	query += `
		WHERE b.deleted_at IS NULL`

	// Применяем те же фильтры
	if filters != nil {
		// Фильтр по региону
		if filters.RegionID != nil {
			query += ` AND JSON_CONTAINS(b.region, ?)`
			args = append(args, fmt.Sprintf("%d", *filters.RegionID))
		}

		// Фильтр по городу
		if filters.CityID != nil {
			query += ` AND b.city_id = UUID_TO_BIN(?)`
			args = append(args, *filters.CityID)
		}

		// Фильтр по типу
		if filters.Type != nil {
			query += ` AND b.type = ?`
			args = append(args, *filters.Type)
		}

		// Фильтр по целевым группам
		if len(filters.TargetGroups) > 0 {
			query += ` AND (`
			for i, group := range filters.TargetGroups {
				if i > 0 {
					query += ` OR `
				}
				query += `JSON_CONTAINS(b.target_group_ids, ?)`
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
				query += `JSON_CONTAINS(b.tags, ?)`
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
				query += `b.category = ?`
				args = append(args, category)
			}
			query += `)`
		}

		// Фильтр по датам
		if filters.DateFrom != nil {
			query += ` AND b.valid_to >= ?`
			args = append(args, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			query += ` AND b.valid_from <= ?`
			args = append(args, *filters.DateTo)
		}

		// Текстовый поиск с Full-Text Search
		if filters.Search != nil && *filters.Search != "" {
			if filters.SearchMode == "boolean" {
				query += ` AND MATCH(b.title, b.description) AGAINST(? IN BOOLEAN MODE)`
			} else {
				query += ` AND MATCH(b.title, b.description) AGAINST(? IN NATURAL LANGUAGE MODE)`
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

func (r *benefitRepository) GetFilterStats(ctx context.Context, filters *BenefitFilters) (*FilterStats, error) {
	// Создаем базовый WHERE clause для фильтров (без фильтров по категориям и уровням)
	baseQuery := `FROM benefit b`
	baseArgs := []interface{}{}

	// JOIN с таблицей favorite если нужно фильтровать по избранным
	if filters != nil && filters.UserID != nil {
		baseQuery += `
		INNER JOIN favorite f ON b.id = f.benefit_id 
			AND f.user_id = UUID_TO_BIN(?) 
			AND f.deleted_at IS NULL`
		baseArgs = append(baseArgs, *filters.UserID)
	}

	baseQuery += `
		WHERE b.deleted_at IS NULL`

	// Применяем фильтры (кроме категорий и типов, так как мы их считаем)
	if filters != nil {
		// Фильтр по региону
		if filters.RegionID != nil {
			baseQuery += ` AND JSON_CONTAINS(b.region, ?)`
			baseArgs = append(baseArgs, fmt.Sprintf("%d", *filters.RegionID))
		}

		// Фильтр по городу
		if filters.CityID != nil {
			baseQuery += ` AND b.city_id = UUID_TO_BIN(?)`
			baseArgs = append(baseArgs, *filters.CityID)
		}

		// Фильтр по целевым группам
		if len(filters.TargetGroups) > 0 {
			baseQuery += ` AND (`
			for i, group := range filters.TargetGroups {
				if i > 0 {
					baseQuery += ` OR `
				}
				baseQuery += `JSON_CONTAINS(b.target_group_ids, ?)`
				baseArgs = append(baseArgs, fmt.Sprintf(`"%s"`, group))
			}
			baseQuery += `)`
		}

		// Фильтр по тегам
		if len(filters.Tags) > 0 {
			baseQuery += ` AND (`
			for i, tag := range filters.Tags {
				if i > 0 {
					baseQuery += ` OR `
				}
				baseQuery += `JSON_CONTAINS(b.tags, ?)`
				baseArgs = append(baseArgs, fmt.Sprintf(`"%s"`, tag))
			}
			baseQuery += `)`
		}

		// Фильтр по датам
		if filters.DateFrom != nil {
			baseQuery += ` AND b.valid_to >= ?`
			baseArgs = append(baseArgs, *filters.DateFrom)
		}
		if filters.DateTo != nil {
			baseQuery += ` AND b.valid_from <= ?`
			baseArgs = append(baseArgs, *filters.DateTo)
		}

		// Текстовый поиск
		if filters.Search != nil && *filters.Search != "" {
			if filters.SearchMode == "boolean" {
				baseQuery += ` AND MATCH(b.title, b.description) AGAINST(? IN BOOLEAN MODE)`
			} else {
				baseQuery += ` AND MATCH(b.title, b.description) AGAINST(? IN NATURAL LANGUAGE MODE)`
			}
			baseArgs = append(baseArgs, *filters.Search)
		}
	}

	// Запрос для получения статистики по категориям
	categoriesQuery := `
		SELECT 
			COALESCE(b.category, 'unknown') as category, 
			COUNT(*) as count
		` + baseQuery + `
		GROUP BY b.category`

	type categoryResult struct {
		Category string `db:"category"`
		Count    int64  `db:"count"`
	}

	var categoryResults []categoryResult
	err := r.db.SelectContext(ctx, &categoryResults, categoriesQuery, baseArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}

	// Запрос для получения статистики по уровням (типам)
	levelsQuery := `
		SELECT 
			b.type as level, 
			COUNT(*) as count
		` + baseQuery + `
		GROUP BY b.type`

	type levelResult struct {
		Level string `db:"level"`
		Count int64  `db:"count"`
	}

	var levelResults []levelResult
	err = r.db.SelectContext(ctx, &levelResults, levelsQuery, baseArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get level stats: %w", err)
	}

	// Формируем результат
	stats := &FilterStats{
		Categories: make(map[string]int64),
		Levels:     make(map[string]int64),
	}

	for _, cr := range categoryResults {
		stats.Categories[cr.Category] = cr.Count
	}

	for _, lr := range levelResults {
		stats.Levels[lr.Level] = lr.Count
	}

	return stats, nil
}
