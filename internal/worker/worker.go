package worker

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/service"
	emailProvider "github.com/vibe-gaming/backend/pkg/email"
)

type Workers struct {
	EmailSender
}

type Deps struct {
	Redis         redis.UniversalClient
	Services      *service.Services
	EmailProvider emailProvider.Sender
	Config        *config.Config
}

type EmailSender interface {
	SendUserVerificationEmail(ctx context.Context, email string, verificationCode string) error
}

func NewWorkers(deps Deps) *Workers {
	return &Workers{
		EmailSender: newEmailSender(deps.EmailProvider, deps.Config.Email),
	}
}
