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
	"github.com/vibe-gaming/backend/internal/queue/client"
	"github.com/vibe-gaming/backend/internal/queue/task"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/hash"
	"github.com/vibe-gaming/backend/pkg/logger"
	"github.com/vibe-gaming/backend/pkg/otp"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

type userService struct {
	userRepository           repository.Users
	refreshSessionRepository repository.RefreshSession
	cityRepository           repository.Cities
	userDocumentRepository   repository.UserDocument
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
	userDocumentRepository repository.UserDocument,
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
		userDocumentRepository:   userDocumentRepository,
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
	user, err := s.userRepository.GetOneByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id failed: %w", err)
	}

	documents, err := s.userDocumentRepository.GetByUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user documents by user id failed: %w", err)
	}

	user.Documents = documents

	return user, nil
}

func (s *userService) UpdateUserInfo(ctx context.Context, userID uuid.UUID, cityID uuid.UUID, groups domain.GroupTypeList) error {
	if _, err := s.cityRepository.GetOneByID(ctx, cityID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrCityNotFound
		}
		return fmt.Errorf("get city by id failed: %w", err)
	}

	user, err := s.userRepository.GetOneByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user by id failed: %w", err)
	}

	// TODO: add transaction
	if err := s.userRepository.UpdateRegisteredAt(ctx, userID); err != nil {
		return fmt.Errorf("update registered at failed: %w", err)
	}

	groupsList := make(domain.UserGroupList, len(groups))
	for i, group := range groups {
		for _, userGroup := range user.GroupType {
			if userGroup.Type == group {
				// skip
				continue
			}
		}

		groupsList[i] = domain.UserGroup{
			Type:   group,
			Status: domain.VerificationStatusPending,
		}
	}

	err = s.userRepository.UpdateUserInfo(ctx, userID, cityID, groupsList)
	if err != nil {
		return fmt.Errorf("update user info failed: %w", err)
	}

	// Запускаем асинхронную задачу для проверки социальных групп
	if user.SNILS.Valid && len(groups) > 0 {
		groupTypes := make([]string, len(groups))
		for i, g := range groups {
			groupTypes[i] = string(g)
		}

		asynqClient := client.GetClient(ctx)
		if asynqClient != nil {
			checkTask, err := task.NewCheckSocialGroupTask(userID, user.SNILS.String, groupTypes)
			if err != nil {
				// Логируем ошибку, но не возвращаем её, чтобы не прерывать основной процесс
				logger.Error("failed to create check social group task", zap.Error(err))
			} else {
				if _, err := asynqClient.Enqueue(checkTask); err != nil {
					logger.Error("failed to enqueue check social group task", zap.Error(err))
				}
			}
		}
	}

	return nil
}

func (s *userService) UpdateUserGroups(ctx context.Context, userID uuid.UUID, groups domain.UserGroupList) error {
	return s.userRepository.UpdateUserGroups(ctx, userID, groups)
}

func (s *userService) CreateDocument(ctx context.Context, document *domain.UserDocument) error {
	return s.userDocumentRepository.Create(ctx, document)
}

func (s *userService) GetDocumentsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserDocument, error) {
	return s.userDocumentRepository.GetByUserID(ctx, userID)
}
