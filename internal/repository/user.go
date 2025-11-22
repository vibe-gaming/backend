package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/db"
	"github.com/vibe-gaming/backend/internal/domain"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func newUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetByExternalID(ctx context.Context, externalID string) (*domain.User, error) {
	const query = `
	SELECT id, external_id, first_name, last_name, middle_name, snils, email, phone_number, created_at, updated_at, deleted_at FROM user WHERE external_id = ?;
	`
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, externalID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select from user by esia_oid failed: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
	INSERT INTO user
	(id, external_id, first_name, last_name, middle_name, snils, email, phone_number)
	VALUES(uuid_to_bin(?), ?, ?, ?, ?, ?, ?, ?);
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.ExternalID,
		user.FirstName.String,
		user.LastName.String,
		user.MiddleName.String,
		user.SNILS.String,
		user.Email.String,
		user.PhoneNumber.String,
	)

	if err != nil {
		//nolint:errorlint
		if mysqlError, ok := err.(*mysql.MySQLError); ok && mysqlError.Number == db.DuplicateEntry {
			return domain.ErrDuplicateEntry
		}
		return fmt.Errorf("db insert esia user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNoRowsAffected
	}

	return nil
}

func (r *userRepository) GetOneByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
	SELECT id, external_id, first_name, last_name, middle_name, snils, email, phone_number, city_id, group_type, registered_at, created_at, updated_at, deleted_at FROM user WHERE id = uuid_to_bin(?);
	`
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select from user by id failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdateUserInfo(ctx context.Context, userID uuid.UUID, cityID uuid.UUID, groups domain.UserGroupList) error {
	const query = `
	UPDATE user SET city_id = uuid_to_bin(?), group_type = ? WHERE id = uuid_to_bin(?);
	`
	_, err := r.db.ExecContext(ctx, query, cityID, groups, userID)
	if err != nil {
		return fmt.Errorf("update user by id failed: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateRegisteredAt(ctx context.Context, userID uuid.UUID) error {
	const query = `
	UPDATE user SET registered_at = now() WHERE id = uuid_to_bin(?);
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("update user by id failed: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateUserGroups(ctx context.Context, userID uuid.UUID, groups domain.UserGroupList) error {
	const query = `
	UPDATE user SET group_type = ? WHERE id = uuid_to_bin(?);
	`
	_, err := r.db.ExecContext(ctx, query, groups, userID)
	if err != nil {
		return fmt.Errorf("update user groups by id failed: %w", err)
	}
	return nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	const query = `SELECT COUNT(*) FROM user WHERE deleted_at IS NULL`
	var count int64
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("count users failed: %w", err)
	}
	return count, nil
}

func (r *userRepository) GetUserGroupsStats(ctx context.Context) (map[string]int64, error) {
	const query = `
		SELECT 
			group_type,
			COUNT(*) as count
		FROM (
			SELECT 
				JSON_UNQUOTE(JSON_EXTRACT(u.group_type, CONCAT('$[', t.idx, '].type'))) as group_type
			FROM user u
			CROSS JOIN JSON_TABLE(
				JSON_ARRAY(0, 1, 2, 3, 4, 5, 6, 7, 8, 9),
				'$[*]' COLUMNS (idx INT PATH '$')
			) AS t
			WHERE u.deleted_at IS NULL
				AND JSON_UNQUOTE(JSON_EXTRACT(u.group_type, CONCAT('$[', t.idx, '].status'))) = 'verified'
				AND JSON_UNQUOTE(JSON_EXTRACT(u.group_type, CONCAT('$[', t.idx, '].type'))) IS NOT NULL
		) AS extracted_groups
		WHERE group_type IS NOT NULL AND group_type != ''
		GROUP BY group_type
	`
	
	type groupStat struct {
		GroupType string `db:"group_type"`
		Count     int64  `db:"count"`
	}
	
	var stats []groupStat
	err := r.db.SelectContext(ctx, &stats, query)
	if err != nil {
		return nil, fmt.Errorf("get user groups stats failed: %w", err)
	}
	
	result := make(map[string]int64)
	for _, stat := range stats {
		if stat.GroupType != "" {
			result[stat.GroupType] = stat.Count
		}
	}
	
	return result, nil
}
