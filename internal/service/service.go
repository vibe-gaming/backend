package service

import (
	"context"
	"log/slog"

	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/email"
	"github.com/vibe-gaming/backend/pkg/hash"
	"github.com/vibe-gaming/backend/pkg/otp"

	"github.com/google/uuid"
)

type Services struct {
	Users Users
}

type Deps struct {
	Logger       *slog.Logger
	Config       *config.Config
	Hasher       hash.PasswordHasher
	TokenManager auth.TokenManager
	OtpGenerator otp.Generator
	EmailSender  email.Sender
	Repos        *repository.Repositories
}

func NewServices(deps Deps) *Services {
	emailService := newEmailsService(deps.EmailSender, deps.Config.Email)

	return &Services{
		Users: newUserService(deps.Repos.Users,
			deps.Repos.RefreshSession,
			deps.Hasher,
			deps.TokenManager,
			deps.OtpGenerator,
			emailService,
			deps.Config.Auth,
		),
	}
}

type Emails interface {
	SendUserVerificationEmail(input VerificationEmailInput) error
}

type Users interface {
	Register(ctx context.Context, input *UserRegisterInput) error
	Auth(ctx context.Context, input *UserAuthInput) (*Tokens, error)
	createSession(ctx context.Context, userID *uuid.UUID, userAgent *string, userIP *string) (*Tokens, error)
}
