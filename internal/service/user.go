package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/hash"
	"github.com/vibe-gaming/backend/pkg/otp"

	"github.com/google/uuid"
)

type userService struct {
	userRepository           repository.Users
	refreshSessionRepository repository.RefreshSession
	hasher                   hash.PasswordHasher
	tokenManager             auth.TokenManager
	otpGenerator             otp.Generator
	authConfig               config.AuthConfig
	config                   *config.Config
}

func newUserService(userRepository repository.Users,
	refreshSessionRepository repository.RefreshSession,
	hasher hash.PasswordHasher,
	tokenManager auth.TokenManager,
	otpGenerator otp.Generator,
	authConfig config.AuthConfig,
	config *config.Config,
) *userService {
	return &userService{
		userRepository:           userRepository,
		refreshSessionRepository: refreshSessionRepository,
		hasher:                   hasher,
		tokenManager:             tokenManager,
		otpGenerator:             otpGenerator,
		authConfig:               authConfig,
		config:                   config,
	}
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

// AuthESIA выполняет авторизацию пользователя через ESIA OAuth
func (s *userService) AuthESIA(ctx context.Context, code string, userAgent string, userIP string) (*Tokens, error) {
	// Создаем ESIA клиент
	esiaClient := esia.NewClient(s.config.ESIA)

	// Обменять код на токен
	tokenResp, err := esiaClient.ExchangeCodeForToken(code)
	if err != nil {
		return nil, fmt.Errorf("exchange code for token failed: %w", err)
	}

	// Получить информацию о пользователе
	userInfo, err := esiaClient.GetUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	// Проверить, существует ли пользователь с таким ESIA OID
	existingUser, err := s.userRepository.GetByESIAOID(ctx, userInfo.OID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("get user by esia oid failed: %w", err)
	}

	var userID uuid.UUID

	if existingUser == nil {
		// Пользователь не найден - создаем нового
		userID, err = uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("generate user id failed: %w", err)
		}

		// Формируем логин из имени и фамилии
		login := fmt.Sprintf("%s_%s", userInfo.FirstName, userInfo.LastName)

		newUser := &domain.User{
			ID:    userID,
			Login: login,
			ESIAOID: sql.NullString{
				String: userInfo.OID,
				Valid:  true,
			},
			ESIAFirstName: sql.NullString{
				String: userInfo.FirstName,
				Valid:  userInfo.FirstName != "",
			},
			ESIALastName: sql.NullString{
				String: userInfo.LastName,
				Valid:  userInfo.LastName != "",
			},
			ESIAMiddleName: sql.NullString{
				String: userInfo.MiddleName,
				Valid:  userInfo.MiddleName != "",
			},
			ESIASNILS: sql.NullString{
				String: userInfo.SNILS,
				Valid:  userInfo.SNILS != "",
			},
			ESIAEmail: sql.NullString{
				String: userInfo.Email,
				Valid:  userInfo.Email != "",
			},
			ESIAMobile: sql.NullString{
				String: userInfo.Mobile,
				Valid:  userInfo.Mobile != "",
			},
		}

		if userInfo.Email != "" {
			newUser.Email = userInfo.Email
		}

		if err := s.userRepository.CreateESIAUser(ctx, newUser); err != nil {
			return nil, fmt.Errorf("create esia user failed: %w", err)
		}
	} else {
		// Пользователь уже существует
		userID = existingUser.ID
	}

	// Создать сессию для пользователя
	tokens, err := s.createSession(ctx, &userID, &userAgent, &userIP)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}

	return tokens, nil
}
