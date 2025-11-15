package service

import (
	"context"

	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/hash"
	"github.com/vibe-gaming/backend/pkg/otp"

	"github.com/google/uuid"
)

type Services struct {
	Users    Users
	Benefits Benefits
	Cities   Cities
}

type Deps struct {
	Config       *config.Config
	Hasher       hash.PasswordHasher
	TokenManager auth.TokenManager
	OtpGenerator otp.Generator
	Repos        *repository.Repositories
	EsiaClient   *esia.Client
}

func NewServices(deps Deps) *Services {
	return &Services{
		Users: newUserService(deps.Repos.Users,
			deps.Repos.RefreshSession,
			deps.Repos.Cities,
			deps.Hasher,
			deps.TokenManager,
			deps.OtpGenerator,
			deps.EsiaClient,
			deps.Config.Auth,
			deps.Config,
		),
		Benefits: newBenefitService(deps.Repos.Benefits),
		Cities:   newCityService(deps.Repos.Cities),
	}
}

type Users interface {
	Auth(ctx context.Context, code string, userAgent string, userIP string) (*Tokens, error)
	createSession(ctx context.Context, userID *uuid.UUID, userAgent *string, userIP *string) (*Tokens, error)
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUserInfo(ctx context.Context, userID uuid.UUID, cityID uuid.UUID, groupType domain.GroupTypeList) error
}

type Cities interface {
	GetAll(ctx context.Context) ([]domain.City, error)
}

type Benefits interface {
	GetAll(ctx context.Context, page, limit int, filters *repository.BenefitFilters) ([]*domain.Benefit, int64, error)
	GetByID(ctx context.Context, id string) (*domain.Benefit, error)
}
