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

type cityRepository struct {
	db *sqlx.DB
}

func newCityRepository(db *sqlx.DB) *cityRepository {
	return &cityRepository{
		db: db,
	}
}

func (r *cityRepository) GetOneByID(ctx context.Context, id uuid.UUID) (*domain.City, error) {
	const query = `
	SELECT id, region_id, name, created_at, updated_at, deleted_at FROM city WHERE id = uuid_to_bin(?);
	`
	var city domain.City
	if err := r.db.GetContext(ctx, &city, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select from city by id failed: %w", err)
	}
	return &city, nil
}

func (r *cityRepository) GetAll(ctx context.Context) ([]domain.City, error) {
	const query = `
	SELECT id, region_id, name, created_at, updated_at, deleted_at FROM city ORDER BY name ASC;
	`
	var cities []domain.City
	if err := r.db.SelectContext(ctx, &cities, query); err != nil {
		return nil, fmt.Errorf("select from city by id failed: %w", err)
	}
	return cities, nil
}
