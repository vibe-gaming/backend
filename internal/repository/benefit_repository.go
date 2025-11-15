package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
)

type BenefitRepository interface {
	Create(ctx context.Context, benefit *domain.Benefit) error
	GetByID(ctx context.Context, id string) (*domain.Benefit, error)
	GetAll(ctx context.Context) ([]*domain.Benefit, error)
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
	INSERT INTO benefit (id, title, description, valid_from, valid_to, created_at, updated_at, deleted_at, type, target_group_ids, longitude, latitude, city_id, region_id, requirement, how_to_use, source_url)
	VALUES (uuid_to_bin(?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, benefit.ID, benefit.Title, benefit.Description, benefit.ValidFrom, benefit.ValidTo, benefit.CreatedAt, benefit.UpdatedAt, benefit.DeletedAt, benefit.Type, benefit.TargetGroupIDs, benefit.Longitude, benefit.Latitude, benefit.CityID, benefit.RegionID, benefit.Requirement, benefit.HowToUse, benefit.SourceURL)
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
			bin_to_uuid(region_id) as region_id,
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

func (r *benefitRepository) GetAll(ctx context.Context) ([]*domain.Benefit, error) {
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
			bin_to_uuid(region_id) as region_id,
			requirment,
			how_to_use,
			source_url
		FROM benefit
		WHERE deleted_at IS NULL
	`
	var benefits []*domain.Benefit
	err := r.db.SelectContext(ctx, &benefits, query)
	if err != nil {
		return nil, err
	}
	return benefits, nil
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
			region_id = uuid_to_bin(?),
			requirment = ?,
			how_to_use = ?,
			source_url = ?
		WHERE id = uuid_to_bin(?)
	`
	_, err := r.db.ExecContext(ctx, query, benefit.Title, benefit.Description, benefit.ValidFrom, benefit.ValidTo, benefit.UpdatedAt, benefit.DeletedAt, benefit.Type, benefit.TargetGroupIDs, benefit.Longitude, benefit.Latitude, benefit.CityID, benefit.RegionID, benefit.Requirement, benefit.HowToUse, benefit.SourceURL, benefit.ID)
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
