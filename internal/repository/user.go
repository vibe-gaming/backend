package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vibe-gaming/backend/internal/db"
	"github.com/vibe-gaming/backend/internal/domain"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
				INSERT INTO user (id, login, password, email) 
				VALUES(uuid_to_bin(?), ?, ?, ?);
				`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Login, user.Password, user.Email)

	if err != nil {
		//nolint:errorlint
		if mysqlError, ok := err.(*mysql.MySQLError); ok && mysqlError.Number == db.DuplicateEntry {
			return domain.ErrDuplicateEntry
		}
		return fmt.Errorf("db insert user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByCredentials(ctx context.Context, email string, password string) (*uuid.UUID, error) {
	const query = `
				SELECT id FROM user
				WHERE email = ?
				AND password = ?
				`
	var ID uuid.UUID
	if err := r.db.GetContext(ctx, &ID, query, email, password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select from user failed: %w", err)
	}

	return &ID, nil
}

func (r *userRepository) GetByESIAOID(ctx context.Context, esiaOID string) (*domain.User, error) {
	const query = `
				SELECT id, login, password, email, esia_oid, esia_first_name, esia_last_name, 
				       esia_middle_name, esia_snils, esia_email, esia_mobile,
				       created_at, updated_at, deleted_at
				FROM user
				WHERE esia_oid = ?
				`
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, esiaOID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select from user by esia_oid failed: %w", err)
	}

	return &user, nil
}

func (r *userRepository) CreateESIAUser(ctx context.Context, user *domain.User) error {
	const query = `
				INSERT INTO user (id, login, email, esia_oid, esia_first_name, esia_last_name, 
				                 esia_middle_name, esia_snils, esia_email, esia_mobile) 
				VALUES(uuid_to_bin(?), ?, ?, ?, ?, ?, ?, ?, ?, ?);
				`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Login,
		user.Email,
		user.ESIAOID,
		user.ESIAFirstName,
		user.ESIALastName,
		user.ESIAMiddleName,
		user.ESIASNILS,
		user.ESIAEmail,
		user.ESIAMobile,
	)

	if err != nil {
		//nolint:errorlint
		if mysqlError, ok := err.(*mysql.MySQLError); ok && mysqlError.Number == db.DuplicateEntry {
			return domain.ErrDuplicateEntry
		}
		return fmt.Errorf("db insert esia user: %w", err)
	}

	return nil
}
