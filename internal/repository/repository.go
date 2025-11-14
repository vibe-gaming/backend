package repository

import (
	"context"

	"github.com/vibe-gaming/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Users          Users
	RefreshSession RefreshSession
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Users:          newUserRepository(db),
		RefreshSession: newRefreshSessionRepository(db),
	}
}

type Users interface {
	Create(ctx context.Context, user *domain.User) error
	GetByCredentials(ctx context.Context, email string, password string) (*uuid.UUID, error)
}

type RefreshSession interface {
	Create(ctx context.Context, session *domain.RefreshSession) error
}
