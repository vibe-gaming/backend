package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

type BenefitFilters struct {
	RegionID            *int
	CityID              *string
	Types               []string // Типы льгот для фильтрации (federal, regional, commercial) - OR логика
	TargetGroups        []string
	Tags                []string
	Categories          []string
	DateFrom            *string
	DateTo              *string
	Search              *string
	SearchMode          string   // "natural" или "boolean"
	SortBy              string   // "created_at", "views", "updated_at"
	Order               string   // "asc", "desc"
	UserID              *string  // UUID пользователя для получения информации об избранном
	FilterFavoritesOnly *bool    // Фильтровать только избранные (favorites=true)
	FilterByUserGroups  *bool    // Фильтровать по группам пользователя
	UserGroupTypes      []string // Подтвержденные группы пользователя для фильтрации
}

type UserBenefitsStats struct {
	TotalBenefits  int64 `json:"total_benefits"`
	TotalFavorites int64 `json:"total_favorites"`
}

type BenefitRepository interface {
	Create(ctx context.Context, benefit *domain.Benefit) error
	GetByID(ctx context.Context, id string, userID *string) (*domain.Benefit, error)
	GetAll(ctx context.Context, limit, offset int, filters *BenefitFilters) ([]*domain.Benefit, error)
	Count(ctx context.Context, filters *BenefitFilters) (int64, error)
	CountAvailableForUser(ctx context.Context, targetGroups []string) (int64, error)
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
func (r *benefitRepository) GetByID(ctx context.Context, id string, userID *string) (*domain.Benefit, error) {
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
			b.views,
			bin_to_uuid(b.organization_id) as organization_id`

	args := []interface{}{}

	// Добавляем поле is_favorite через LEFT JOIN с favorite
	if userID != nil {
		query += `,
			CASE WHEN f.id IS NOT NULL THEN 1 ELSE 0 END as is_favorite`
	} else {
		query += `,
			0 as is_favorite`
	}

	query += `
		FROM benefit b`

	// LEFT JOIN с таблицей favorite для получения информации об избранном
	if userID != nil {
		query += `
		LEFT JOIN favorite f ON b.id = f.benefit_id 
			AND f.user_id = UUID_TO_BIN(?)
			AND f.deleted_at IS NULL`
		args = append(args, *userID)
	}

	query += `
		WHERE b.id = uuid_to_bin(?) AND b.deleted_at IS NULL`
	args = append(args, id)

	var benefit domain.Benefit
	err := r.db.GetContext(ctx, &benefit, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	logger.Info("benefit organization id", zap.Any("organization_id", benefit.OrganizationID))
	return &benefit, nil
}

func (r *benefitRepository) GetAll(ctx context.Context, limit, offset int, filters *BenefitFilters) ([]*domain.Benefit, error) {
	// Проверяем, есть ли поисковый запрос для расчета релевантности
	hasSearch := filters != nil && filters.Search != nil && *filters.Search != ""

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
			b.views,
			b.organization_id`

	// Добавляем поле is_favorite через LEFT JOIN с favorite
	if filters != nil && filters.UserID != nil {
		query += `,
			CASE WHEN f.id IS NOT NULL THEN 1 ELSE 0 END as is_favorite`
	} else {
		query += `,
			0 as is_favorite`
	}

	// Добавляем поле релевантности, если есть поисковый запрос
	if hasSearch {
		// Добавляем score релевантности - будет использоваться для сортировки
		// Не добавляем в SELECT, т.к. структура Benefit не содержит это поле
		// Будем использовать MATCH() AGAINST() напрямую в ORDER BY
	}

	query += `
		FROM benefit b`

	args := []interface{}{}

	// LEFT JOIN с таблицей favorite для получения информации об избранном
	if filters != nil && filters.UserID != nil {
		query += `
		LEFT JOIN favorite f ON b.id = f.benefit_id 
			AND f.user_id = UUID_TO_BIN(?)
			AND f.deleted_at IS NULL`
		args = append(args, *filters.UserID)
	}

	query += `
		WHERE b.deleted_at IS NULL`

	// Если нужно фильтровать только избранные
	if filters != nil && filters.FilterFavoritesOnly != nil && *filters.FilterFavoritesOnly && filters.UserID != nil {
		query += ` AND f.id IS NOT NULL`
	}

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

		// Фильтр по типам (хотя бы один тип должен совпадать) - OR логика
		if len(filters.Types) > 0 {
			query += ` AND (`
			for i, benefitType := range filters.Types {
				if i > 0 {
					query += ` OR `
				}
				query += `b.type = ?`
				args = append(args, benefitType)
			}
			query += `)`
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

		// Фильтр по группам пользователя (показать только доступные пользователю льготы)
		if filters.FilterByUserGroups != nil && *filters.FilterByUserGroups {
			if len(filters.UserGroupTypes) > 0 {
				query += ` AND (`
				for i, group := range filters.UserGroupTypes {
					if i > 0 {
						query += ` OR `
					}
					query += `JSON_CONTAINS(b.target_group_ids, ?)`
					args = append(args, fmt.Sprintf(`"%s"`, group))
				}
				query += `)`
			} else {
				// Если фильтр включен, но у пользователя нет групп - вернуть пустой результат
				query += ` AND FALSE`
			}
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
	var orderClause string

	// Если есть поисковый запрос, сортируем по релевантности
	if hasSearch {
		// Используем MATCH() AGAINST() для расчета релевантности
		// В Boolean mode релевантность тоже работает, но менее точно
		// Поэтому дублируем запрос в ORDER BY для получения score
		if filters.SearchMode == "boolean" {
			orderClause = fmt.Sprintf("MATCH(b.title, b.description) AGAINST(? IN BOOLEAN MODE) DESC")
			args = append(args, *filters.Search)
		} else {
			orderClause = fmt.Sprintf("MATCH(b.title, b.description) AGAINST(? IN NATURAL LANGUAGE MODE) DESC")
			args = append(args, *filters.Search)
		}
	} else {
		// Обычная сортировка без поиска
		if filters != nil && filters.SortBy != "" {
			// Если указана сортировка, используем её
			orderBy := "b.created_at"
			orderDir := "DESC"

			// Определяем поле для сортировки
			switch filters.SortBy {
			case "views":
				orderBy = "b.views"
			case "updated_at":
				orderBy = "b.updated_at"
			case "created_at":
				orderBy = "b.created_at"
			case "relevance":
				// relevance без поиска не имеет смысла, используем created_at
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

			orderClause = fmt.Sprintf("%s %s", orderBy, orderDir)
		} else {
			// Если сортировка не указана, сортируем сначала по типу (federal, regional, commercial), затем по created_at
			orderClause = `CASE b.type 
				WHEN 'federal' THEN 1 
				WHEN 'regional' THEN 2 
				WHEN 'commercial' THEN 3 
				ELSE 4 
			END ASC, b.created_at DESC`
		}
	}

	query += fmt.Sprintf(`
		ORDER BY %s
		LIMIT ? OFFSET ?`, orderClause)

	args = append(args, limit, offset)

	var benefits []*domain.Benefit
	err := r.db.SelectContext(ctx, &benefits, query, args...)
	if err != nil {

		return nil, err
	}

	organizationRepository := NewOrganizationRepository(r.db)
	for _, benefit := range benefits {

		if benefit.OrganizationID != nil {
			organization, err := organizationRepository.GetByID(ctx, benefit.OrganizationID.String())
			if err != nil {
				return nil, err
			}

			benefit.Organization = organization
		}

	}

	return benefits, nil
}

func (r *benefitRepository) Count(ctx context.Context, filters *BenefitFilters) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM benefit b`

	args := []interface{}{}

	// LEFT JOIN с таблицей favorite для получения информации об избранном
	if filters != nil && filters.UserID != nil {
		query += `
		LEFT JOIN favorite f ON b.id = f.benefit_id 
			AND f.user_id = UUID_TO_BIN(?)
			AND f.deleted_at IS NULL`
		args = append(args, *filters.UserID)
	}

	query += `
		WHERE b.deleted_at IS NULL`

	// Если нужно фильтровать только избранные
	if filters != nil && filters.FilterFavoritesOnly != nil && *filters.FilterFavoritesOnly && filters.UserID != nil {
		query += ` AND f.id IS NOT NULL`
	}

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

		// Фильтр по типам (хотя бы один тип должен совпадать) - OR логика
		if len(filters.Types) > 0 {
			query += ` AND (`
			for i, benefitType := range filters.Types {
				if i > 0 {
					query += ` OR `
				}
				query += `b.type = ?`
				args = append(args, benefitType)
			}
			query += `)`
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

		// Фильтр по группам пользователя (показать только доступные пользователю льготы)
		if filters.FilterByUserGroups != nil && *filters.FilterByUserGroups {
			if len(filters.UserGroupTypes) > 0 {
				query += ` AND (`
				for i, group := range filters.UserGroupTypes {
					if i > 0 {
						query += ` OR `
					}
					query += `JSON_CONTAINS(b.target_group_ids, ?)`
					args = append(args, fmt.Sprintf(`"%s"`, group))
				}
				query += `)`
			} else {
				// Если фильтр включен, но у пользователя нет групп - вернуть пустой результат
				query += ` AND FALSE`
			}
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

	// LEFT JOIN с таблицей favorite для получения информации об избранном
	if filters != nil && filters.UserID != nil {
		baseQuery += `
		LEFT JOIN favorite f ON b.id = f.benefit_id 
			AND f.user_id = UUID_TO_BIN(?)
			AND f.deleted_at IS NULL`
		baseArgs = append(baseArgs, *filters.UserID)
	}

	baseQuery += `
		WHERE b.deleted_at IS NULL`

	// Если нужно фильтровать только избранные
	if filters != nil && filters.FilterFavoritesOnly != nil && *filters.FilterFavoritesOnly && filters.UserID != nil {
		baseQuery += ` AND f.id IS NOT NULL`
	}

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

		// Фильтр по группам пользователя (показать только доступные пользователю льготы)
		if filters.FilterByUserGroups != nil && *filters.FilterByUserGroups {
			if len(filters.UserGroupTypes) > 0 {
				baseQuery += ` AND (`
				for i, group := range filters.UserGroupTypes {
					if i > 0 {
						baseQuery += ` OR `
					}
					baseQuery += `JSON_CONTAINS(b.target_group_ids, ?)`
					baseArgs = append(baseArgs, fmt.Sprintf(`"%s"`, group))
				}
				baseQuery += `)`
			} else {
				// Если фильтр включен, но у пользователя нет групп - вернуть пустой результат
				baseQuery += ` AND FALSE`
			}
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

// CountAvailableForUser подсчитывает количество льгот, доступных для пользователя
// на основе его групп (используется OR логика - хотя бы одна группа должна совпадать)
func (r *benefitRepository) CountAvailableForUser(ctx context.Context, targetGroups []string) (int64, error) {
	if len(targetGroups) == 0 {
		// Если у пользователя нет групп, возвращаем 0
		return 0, nil
	}

	query := `
		SELECT COUNT(*) 
		FROM benefit b
		WHERE b.deleted_at IS NULL`

	args := []interface{}{}

	// Добавляем условие для групп (OR логика - хотя бы одна группа должна совпадать)
	query += ` AND (`
	for i, group := range targetGroups {
		if i > 0 {
			query += ` OR `
		}
		query += `JSON_CONTAINS(b.target_group_ids, ?)`
		args = append(args, fmt.Sprintf(`"%s"`, group))
	}
	query += `)`

	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
