package service

import (
	"context"

	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/repository"
)

type BenefitService struct {
	benefitRepository repository.BenefitRepository
}

func newBenefitService(benefitRepository repository.BenefitRepository) *BenefitService {
	return &BenefitService{
		benefitRepository: benefitRepository,
	}
}

func (s *BenefitService) GetAll(ctx context.Context) ([]*domain.Benefit, error) {
	return s.benefitRepository.GetAll(ctx)
}

func (s *BenefitService) GetByID(ctx context.Context, id string) (*domain.Benefit, error) {
	return s.benefitRepository.GetByID(ctx, id)
}
