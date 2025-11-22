package service

import (
	"context"
	"time"

	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/repository"
)

type organizationService struct {
	organizationRepository repository.OrganizationRepository
}

func newOrganizationService(organizationRepository repository.OrganizationRepository) *organizationService {
	return &organizationService{
		organizationRepository: organizationRepository,
	}
}

func (s *organizationService) Create(ctx context.Context, organization *domain.Organization) error {
	now := time.Now()
	if organization.CreatedAt.IsZero() {
		organization.CreatedAt = now
	}
	if organization.UpdatedAt.IsZero() {
		organization.UpdatedAt = now
	}
	return s.organizationRepository.Create(ctx, organization)
}

func (s *organizationService) Update(ctx context.Context, organization *domain.Organization) error {
	organization.UpdatedAt = time.Now()
	return s.organizationRepository.Update(ctx, organization)
}

func (s *organizationService) Delete(ctx context.Context, id string) error {
	return s.organizationRepository.Delete(ctx, id)
}

func (s *organizationService) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	return s.organizationRepository.GetByID(ctx, id)
}

func (s *organizationService) GetAll(ctx context.Context) ([]domain.Organization, error) {
	return s.organizationRepository.GetAll(ctx)
}

func (s *organizationService) GetAllByCityID(ctx context.Context, cityID string) ([]domain.Organization, error) {
	return s.organizationRepository.GetAllByCityID(ctx, cityID)
}

