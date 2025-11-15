package repository

import (
	"context"

	"github.com/vibe-gaming/backend/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Users          Users
	RefreshSession RefreshSession
	Benefits       BenefitRepository
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Users:          newUserRepository(db),
		RefreshSession: newRefreshSessionRepository(db),
		Benefits:       NewBenefitRepository(db),
	}
}

type Users interface {
	GetByExternalID(ctx context.Context, esiaOID string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

type RefreshSession interface {
	Create(ctx context.Context, session *domain.RefreshSession) error
}
