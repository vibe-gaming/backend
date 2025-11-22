package service

import (
	"context"

	"github.com/vibe-gaming/backend/internal/repository"
)

type favoriteService struct {
	favoriteRepository repository.FavoriteRepository
}

func newFavoriteService(favoriteRepository repository.FavoriteRepository) *favoriteService {
	return &favoriteService{
		favoriteRepository: favoriteRepository,
	}
}

func (s *favoriteService) GetTotalCount(ctx context.Context) (int64, error) {
	return s.favoriteRepository.GetTotalCount(ctx)
}

