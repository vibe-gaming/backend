package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
)

type userVerificationRepository struct {
	db *sqlx.DB
}

func newUserVerificationRepository(db *sqlx.DB) *userVerificationRepository {
	return &userVerificationRepository{
		db: db,
	}
}

func (r *userVerificationRepository) Create(ctx context.Context, userVerification *domain.UserVerification) error {
	const op = "repository.userVerification.Create"

	const query = `
    INSERT INTO user_verification (id, user_id, email, code)
    VALUES (uuid_to_bin(:id), uuid_to_bin(:user_id), :email, :code)
    `

	res, err := r.db.NamedExecContext(ctx, query, userVerification)
	if err != nil {
		return fmt.Errorf("%s: insert user verification failed: %w", op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected failed: %w", op, err)
	}

	if rows != 1 {
		return fmt.Errorf("%s: expected 1 row affected, got %d", op, rows)
	}

	return nil
}

func (r *userVerificationRepository) GetOneByID(ctx context.Context, id uuid.UUID) (*domain.UserVerification, error) {
	const op = "repository.userVerification.GetOneByID"

	const query = `
    SELECT id, user_id, email, code, attempts, confirmed, confirmed_at, created_at, updated_at, deleted_at
    FROM user_verification
    WHERE id = uuid_to_bin(?)
    `

	var userVerification domain.UserVerification
	if err := r.db.GetContext(ctx, &userVerification, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%s: select user verification failed: %w", op, err)
	}

	return &userVerification, nil
}

func (r *userVerificationRepository) UpdateConfirmedByIDWithTx(ctx context.Context, tx *sqlx.Tx, userVerification *domain.UserVerification) error {
	const op = "repository.userVerification.UpdateConfirmedByID"

	const query = `
    UPDATE user_verification
    SET confirmed = ?, confirmed_at = ?
    WHERE id = uuid_to_bin(?)
    `

	res, err := tx.ExecContext(ctx, query, userVerification.Confirmed, userVerification.ConfirmedAt, userVerification.ID)
	if err != nil {
		return fmt.Errorf("%s: update user_verification failed: %w", op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected failed: %w", op, err)
	}

	if rows != 1 {
		return fmt.Errorf("%s: expected 1 row affected, got %d", op, rows)
	}

	return nil
}
