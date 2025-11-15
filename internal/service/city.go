package service

import (
	"context"

	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/repository"
)

type cityService struct {
	cityRepository repository.Cities
}

func newCityService(cityRepository repository.Cities) *cityService {
	return &cityService{
		cityRepository: cityRepository,
	}
}

func (s *cityService) GetAll(ctx context.Context) ([]domain.City, error) {
	return s.cityRepository.GetAll(ctx)
}
