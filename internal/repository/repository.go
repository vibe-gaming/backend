package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Users          Users
	RefreshSession RefreshSession
	Benefits       BenefitRepository
	Cities         Cities
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Users:          newUserRepository(db),
		RefreshSession: newRefreshSessionRepository(db),
		Benefits:       NewBenefitRepository(db),
		Cities:         newCityRepository(db),
	}
}

type Users interface {
	GetByExternalID(ctx context.Context, esiaOID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUserInfo(ctx context.Context, userID uuid.UUID, cityID uuid.UUID, groupType domain.GroupTypeList) error
	UpdateRegisteredAt(ctx context.Context, userID uuid.UUID) error
}

type RefreshSession interface {
	Create(ctx context.Context, session *domain.RefreshSession) error
}

type Cities interface {
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.City, error)
	GetAll(ctx context.Context) ([]domain.City, error)
}
