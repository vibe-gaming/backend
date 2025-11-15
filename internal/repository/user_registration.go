package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
)

type userRegistrationRepository struct {
	db *sqlx.DB
}

func newUserRegistrationRepository(db *sqlx.DB) *userRegistrationRepository {
	return &userRegistrationRepository{
		db: db,
	}
}

func (r *userRegistrationRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.UserRegistration, error) {
	var userRegistration domain.UserRegistration
	const query = "SELECT id, email, phone, password, code, confirmed, confirmed_at, created_at, updated_at, deleted_at FROM user_registration WHERE id = uuid_to_bin(?)"

	if err := r.db.GetContext(ctx, &userRegistration, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select query err: %w", err)
	}

	return &userRegistration, nil
}

func (r *userRegistrationRepository) Create(ctx context.Context, userRegistration *domain.UserRegistration) error {
	const query = "INSERT INTO user_registration (id, email, phone, password, code) VALUES (uuid_to_bin(?), ?, ?, ?, ?, uuid_to_bin(?), ?)"

	_, err := r.db.ExecContext(ctx, query, userRegistration.ID, userRegistration.Email, userRegistration.Phone, userRegistration.Password, userRegistration.Code)
	if err != nil {
		return fmt.Errorf("db insert user registration: %w", err)
	}

	return nil
}

func (r *userRegistrationRepository) Verify(ctx context.Context, id uuid.UUID, code string) error {
	uuidBytes, err := id.MarshalBinary()
	if err != nil {
		return fmt.Errorf("email verify uuid marshal failed: %w", err)
	}

	const query = "UPDATE user_registration SET confirmed=?, confirmed_at=? WHERE id=? and code=? and confirmed=?"
	res, err := r.db.ExecContext(ctx, query, true, time.Now(), uuidBytes, code, false)
	if err != nil {
		return fmt.Errorf("db update user registration: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("db update user registration rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrNoRowsAffected
	}

	return nil
}
