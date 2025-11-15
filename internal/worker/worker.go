package worker

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/service"
	socialgroupchecker "github.com/vibe-gaming/backend/internal/service/social_group_checker"
	emailProvider "github.com/vibe-gaming/backend/pkg/email"
)

type Workers struct {
	EmailSender        EmailSender
	SocialGroupChecker SocialGroupChecker
}

type Deps struct {
	Redis                    redis.UniversalClient
	Services                 *service.Services
	EmailProvider            emailProvider.Sender
	Config                   *config.Config
	SocialGroupCheckerClient *socialgroupchecker.Client
}

type EmailSender interface {
	SendUserVerificationEmail(ctx context.Context, email string, verificationCode string) error
}

type SocialGroupChecker interface {
	CheckGroups(ctx context.Context, snils string, groups []socialgroupchecker.SocialGroup) (*socialgroupchecker.CheckResponse, error)
	CheckAndUpdateUserGroups(ctx context.Context, userID uuid.UUID, snils string, groupTypes []string) error
}

func NewWorkers(deps Deps) *Workers {
	return &Workers{
		EmailSender:        newEmailSender(deps.EmailProvider, deps.Config.Email),
		SocialGroupChecker: newSocialGroupChecker(deps.SocialGroupCheckerClient, deps.Services),
	}
}
