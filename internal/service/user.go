package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/hash"
	"github.com/vibe-gaming/backend/pkg/otp"

	"github.com/google/uuid"
)

type userService struct {
	userRepository             repository.Users
	userRegistrationRepository repository.UserRegistration
	refreshSessionRepository   repository.RefreshSession
	hasher                     hash.PasswordHasher
	tokenManager               auth.TokenManager
	otpGenerator               otp.Generator
	authConfig                 config.AuthConfig
	config                     *config.Config
}

func newUserService(userRepository repository.Users,
	userRegistrationRepository repository.UserRegistration,
	refreshSessionRepository repository.RefreshSession,
	hasher hash.PasswordHasher,
	tokenManager auth.TokenManager,
	otpGenerator otp.Generator,
	authConfig config.AuthConfig,
	config *config.Config,
) *userService {
	return &userService{
		userRepository:             userRepository,
		userRegistrationRepository: userRegistrationRepository,
		refreshSessionRepository:   refreshSessionRepository,
		hasher:                     hasher,
		tokenManager:               tokenManager,
		otpGenerator:               otpGenerator,
		authConfig:                 authConfig,
		config:                     config,
	}
}

type UserRegisterInput struct {
	Phone    string
	Login    string
	Email    string
	Password string
}

func (s *userService) Register(ctx context.Context, input *UserRegisterInput) error {
	panic("not implemented")
}

type UserAuthInput struct {
	Email     string
	Password  string
	UserAgent string
	IP        string
}

func (s *userService) Auth(ctx context.Context, input *UserAuthInput) (*Tokens, error) {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hasher failed: %w", err)
	}

	userID, err := s.userRepository.GetByCredentials(ctx, input.Email, passwordHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get by credentials failed: %w", err)
	}

	tokens, err := s.createSession(ctx, userID, &input.UserAgent, &input.IP)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}

	return tokens, nil
}

type Tokens struct {
	AccessToken  string
	AccessTTL    time.Duration
	RefreshToken uuid.UUID
	RefreshTTL   time.Duration
}

func (s *userService) createSession(ctx context.Context, userID *uuid.UUID, userAgent *string, userIP *string) (*Tokens, error) {
	var (
		res Tokens
		err error
	)

	res.AccessToken, res.AccessTTL, err = s.tokenManager.NewJWT(userID)
	if err != nil {
		return &res, fmt.Errorf("generate access token failed: %w", err)
	}

	res.RefreshToken, res.RefreshTTL, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return &res, fmt.Errorf("generate refresh token failed: %w", err)
	}

	refreshSessionID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate refresh session id failed: %w", err)
	}
	refreshSession := &domain.RefreshSession{
		ID:           refreshSessionID,
		UserID:       *userID,
		RefreshToken: res.RefreshToken,
		UserAgent:    *userAgent,
		IP:           *userIP,
		ExpiresIn:    time.Now().Add(res.RefreshTTL),
	}

	if err := s.refreshSessionRepository.Create(ctx, refreshSession); err != nil {
		return nil, fmt.Errorf("create refresh session failed: %w", err)
	}

	return &res, nil
}

func (s *userService) Verify(ctx context.Context, id uuid.UUID, code string) (*Tokens, error) {
	panic("not implemented")
}
