package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
)

type FavoriteRepository interface {
	GetByUserIDAndBenefitID(ctx context.Context, userID uuid.UUID, benefitID uuid.UUID) (*domain.Favorite, error)
	Create(ctx context.Context, favorite *domain.Favorite) error
	Update(ctx context.Context, favorite *domain.Favorite) error
}

type favoriteRepository struct {
	db *sqlx.DB
}

func NewFavoriteRepository(db *sqlx.DB) FavoriteRepository {
	return &favoriteRepository{
		db: db,
	}
}

func (r *favoriteRepository) GetByUserIDAndBenefitID(ctx context.Context, userID uuid.UUID, benefitID uuid.UUID) (*domain.Favorite, error) {
	const query = `
		SELECT id, user_id, benefit_id, created_at, updated_at, deleted_at FROM favorite WHERE user_id = uuid_to_bin(?) AND benefit_id = uuid_to_bin(?)
	`
	var favorite domain.Favorite
	err := r.db.GetContext(ctx, &favorite, query, userID, benefitID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &favorite, nil
}

func (r *favoriteRepository) Create(ctx context.Context, favorite *domain.Favorite) error {
	const query = `
		INSERT INTO favorite (id, user_id, benefit_id, created_at, updated_at) VALUES (uuid_to_bin(?), uuid_to_bin(?), uuid_to_bin(?), ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, favorite.ID, favorite.UserID, favorite.BenefitID, favorite.CreatedAt, favorite.UpdatedAt)
	if err != nil {
		return fmt.Errorf("db create favorite: %w", err)
	}
	return nil
}

func (r *favoriteRepository) Update(ctx context.Context, favorite *domain.Favorite) error {
	const query = `
		UPDATE favorite SET deleted_at = ?, updated_at = ? WHERE id = uuid_to_bin(?)
	`
	_, err := r.db.ExecContext(ctx, query, favorite.DeletedAt, favorite.UpdatedAt, favorite.ID)
	if err != nil {
		return fmt.Errorf("db update favorite: %w", err)
	}
	return nil
}
