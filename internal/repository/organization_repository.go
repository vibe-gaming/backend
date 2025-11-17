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
		tags
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
