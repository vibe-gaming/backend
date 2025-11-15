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
	cityRepository           repository.Cities
	hasher                   hash.PasswordHasher
	tokenManager             auth.TokenManager
	otpGenerator             otp.Generator
	esiaClient               *esia.Client
	authConfig               config.AuthConfig
	config                   *config.Config
}

func newUserService(userRepository repository.Users,
	refreshSessionRepository repository.RefreshSession,
	cityRepository repository.Cities,
	hasher hash.PasswordHasher,
	tokenManager auth.TokenManager,
	otpGenerator otp.Generator,
	esiaClient *esia.Client,
	authConfig config.AuthConfig,
	config *config.Config,
) *userService {
	return &userService{
		userRepository:           userRepository,
		refreshSessionRepository: refreshSessionRepository,
		cityRepository:           cityRepository,
		hasher:                   hasher,
		tokenManager:             tokenManager,
		otpGenerator:             otpGenerator,
		esiaClient:               esiaClient,
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

// Auth выполняет авторизацию пользователя через ESIA OAuth
func (s *userService) Auth(ctx context.Context, code string, userAgent string, userIP string) (*Tokens, error) {
	// Обменять код на токен
	tokenResp, err := s.esiaClient.ExchangeCodeForToken(code)
	if err != nil {
		return nil, fmt.Errorf("exchange code for token failed: %w", err)
	}

	// Получить информацию о пользователе
	userInfo, err := s.esiaClient.GetUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	// Проверить, существует ли пользователь с таким ESIA OID
	existingUser, err := s.userRepository.GetByExternalID(ctx, userInfo.OID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("get user by external id failed: %w", err)
	}

	var userID uuid.UUID

	if existingUser == nil {
		// Пользователь не найден - создаем нового
		userID, err = uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("generate user id failed: %w", err)
		}

		newUser := &domain.User{
			ID: userID,
			ExternalID: sql.NullString{
				String: userInfo.OID,
				Valid:  true,
			},
			FirstName: sql.NullString{
				String: userInfo.FirstName,
				Valid:  userInfo.FirstName != "",
			},
			LastName: sql.NullString{
				String: userInfo.LastName,
				Valid:  userInfo.LastName != "",
			},
			MiddleName: sql.NullString{
				String: userInfo.MiddleName,
				Valid:  userInfo.MiddleName != "",
			},
			SNILS: sql.NullString{
				String: userInfo.SNILS,
				Valid:  userInfo.SNILS != "",
			},
			Email: sql.NullString{
				String: userInfo.Email,
				Valid:  userInfo.Email != "",
			},
			PhoneNumber: sql.NullString{
				String: userInfo.Mobile,
				Valid:  userInfo.Mobile != "",
			},
		}

		if err := s.userRepository.Create(ctx, newUser); err != nil {
			return nil, fmt.Errorf("create user failed: %w", err)
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

func (s *userService) GetOneByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepository.GetOneByID(ctx, id)
}

func (s *userService) UpdateUserInfo(ctx context.Context, userID uuid.UUID, cityID uuid.UUID, groupType domain.GroupTypeList) error {
	if _, err := s.cityRepository.GetOneByID(ctx, cityID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrCityNotFound
		}
		return fmt.Errorf("get city by id failed: %w", err)
	}

	return s.userRepository.CompleteRegistration(ctx, userID, cityID, groupType)
}
