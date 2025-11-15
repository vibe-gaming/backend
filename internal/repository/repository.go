package repository

import (
	"context"

	"github.com/vibe-gaming/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Users            Users
	UserRegistration UserRegistration
	RefreshSession   RefreshSession
	UserVerification UserVerification
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Users:            newUserRepository(db),
		UserRegistration: newUserRegistrationRepository(db),
		RefreshSession:   newRefreshSessionRepository(db),
		UserVerification: newUserVerificationRepository(db),
	}
}

type Users interface {
	Create(ctx context.Context, user *domain.User) error
	GetByCredentials(ctx context.Context, email string, password string) (*uuid.UUID, error)
}

type UserVerification interface {
	Create(ctx context.Context, userVerification *domain.UserVerification) error
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.UserVerification, error)
	UpdateConfirmedByIDWithTx(ctx context.Context, tx *sqlx.Tx, userVerification *domain.UserVerification) error
}

type UserRegistration interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.UserRegistration, error)
	Create(ctx context.Context, userRegistration *domain.UserRegistration) error
	Verify(ctx context.Context, id uuid.UUID, code string) error
}

type RefreshSession interface {
	Create(ctx context.Context, session *domain.RefreshSession) error
}
