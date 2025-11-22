package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

type OrganizationRepository interface {
	Create(ctx context.Context, organization *domain.Organization) error
	GetByID(ctx context.Context, id string) (*domain.Organization, error)
	GetAll(ctx context.Context) ([]domain.Organization, error)
	GetAllByCityID(ctx context.Context, cityID string) ([]domain.Organization, error)
	Update(ctx context.Context, organization *domain.Organization) error
	Delete(ctx context.Context, id string) error
}

type organizationRepository struct {
	db *sqlx.DB
}

func NewOrganizationRepository(db *sqlx.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) Create(ctx context.Context, organization *domain.Organization) error {
	const query = `
	INSERT INTO organization (id, name, description, created_at, updated_at, deleted_at)
	VALUES (uuid_to_bin(?), ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, organization.ID, organization.Name, organization.Description, organization.CreatedAt, organization.UpdatedAt, organization.DeletedAt)
	if err != nil {
		return fmt.Errorf("db insert organization: %w", err)
	}
	return nil
}

func (r *organizationRepository) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	const query = `
	SELECT 
		BIN_TO_UUID(id) as id, 
		name, 
		description, 
		created_at, 
		updated_at, 
		deleted_at 
	FROM organization 
	WHERE id = uuid_to_bin(?) AND deleted_at IS NULL
	`
	var organization domain.Organization
	err := r.db.GetContext(ctx, &organization, query, id)
	if err != nil {
		logger.Error("failed to get organization", zap.Error(err), zap.String("organization_id", id))
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	const buildingsQuery = `
	SELECT 
		BIN_TO_UUID(id) as id, 
		BIN_TO_UUID(organization_id) as organization_id, 
		created_at, 
		updated_at, 
		deleted_at, 
		address, 
		latitude, 
		longitude, 
		phone_number, 
		start_time, 
		end_time, 
		is_open, 
		tags,
		type
	FROM organization_building 
	WHERE organization_id = uuid_to_bin(?) AND deleted_at IS NULL
	`

	var buildings []domain.OrganizationBuilding
	err = r.db.SelectContext(ctx, &buildings, buildingsQuery, id)
	if err != nil {
		logger.Error("failed to get organization buildings", zap.Error(err), zap.String("organization_id", id))
		return nil, fmt.Errorf("failed to get organization buildings: %w", err)
	}

	logger.Info("organization buildings loaded", zap.Int("count", len(buildings)))
	organization.Buildings = buildings

	return &organization, nil
}

func (r *organizationRepository) GetAll(ctx context.Context) ([]domain.Organization, error) {
	const query = `
	SELECT 
		BIN_TO_UUID(id) as id, 
		name, 
		description, 
		created_at, 
		updated_at, 
		deleted_at 
	FROM organization 
	WHERE deleted_at IS NULL
	ORDER BY name ASC
	`
	var organizations []domain.Organization
	err := r.db.SelectContext(ctx, &organizations, query)
	if err != nil {
		logger.Error("failed to get organizations", zap.Error(err))
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	return organizations, nil
}

func (r *organizationRepository) GetAllByCityID(ctx context.Context, cityID string) ([]domain.Organization, error) {
	const query = `
	SELECT DISTINCT
		BIN_TO_UUID(o.id) as id, 
		o.name, 
		o.description, 
		o.created_at, 
		o.updated_at, 
		o.deleted_at 
	FROM organization o
	INNER JOIN benefit b ON b.organization_id = o.id
	WHERE o.deleted_at IS NULL 
		AND b.deleted_at IS NULL
		AND b.city_id = UUID_TO_BIN(?)
	ORDER BY o.name ASC
	`
	var organizations []domain.Organization
	err := r.db.SelectContext(ctx, &organizations, query, cityID)
	if err != nil {
		logger.Error("failed to get organizations by city", zap.Error(err), zap.String("city_id", cityID))
		return nil, fmt.Errorf("failed to get organizations by city: %w", err)
	}
	return organizations, nil
}

func (r *organizationRepository) Update(ctx context.Context, organization *domain.Organization) error {
	const query = `
	UPDATE organization SET name = ?, description = ?, updated_at = ?, deleted_at = ? WHERE id = uuid_to_bin(?)
	`
	_, err := r.db.ExecContext(ctx, query, organization.Name, organization.Description, organization.UpdatedAt, organization.DeletedAt, organization.ID)
	if err != nil {
		return fmt.Errorf("db update organization: %w", err)
	}
	return nil
}

func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	const query = `
	UPDATE organization SET deleted_at = NOW() WHERE id = uuid_to_bin(?) AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("db delete organization: %w", err)
	}
	return nil
}
