package service

import (
	"context"

	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/hash"
	"github.com/vibe-gaming/backend/pkg/otp"

	"github.com/google/uuid"
)

type Services struct {
	Users Users
}

type Deps struct {
	Config       *config.Config
	Hasher       hash.PasswordHasher
	TokenManager auth.TokenManager
	OtpGenerator otp.Generator
	Repos        *repository.Repositories
}

func NewServices(deps Deps) *Services {
	return &Services{
		Users: newUserService(deps.Repos.Users,
			deps.Repos.RefreshSession,
			deps.Hasher,
			deps.TokenManager,
			deps.OtpGenerator,
			deps.Config.Auth,
			deps.Config,
		),
	}
}

type Users interface {
	AuthESIA(ctx context.Context, code string, userAgent string, userIP string) (*Tokens, error)
	createSession(ctx context.Context, userID *uuid.UUID, userAgent *string, userIP *string) (*Tokens, error)
}
