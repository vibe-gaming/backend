package service

import (
	"context"

	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/repository"
)

// BenefitFilters - псевдоним для удобства использования
type BenefitFilters = repository.BenefitFilters

type BenefitService struct {
	benefitRepository repository.BenefitRepository
}

func newBenefitService(benefitRepository repository.BenefitRepository) *BenefitService {
	return &BenefitService{
		benefitRepository: benefitRepository,
	}
}

func (s *BenefitService) GetAll(ctx context.Context, page, limit int, filters *BenefitFilters) ([]*domain.Benefit, int64, error) {
	// Валидация параметров пагинации
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	benefits, err := s.benefitRepository.GetAll(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.benefitRepository.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return benefits, total, nil
}

func (s *BenefitService) GetByID(ctx context.Context, id string) (*domain.Benefit, error) {
	return s.benefitRepository.GetByID(ctx, id)
}
