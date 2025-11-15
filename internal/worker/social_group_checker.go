package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/service"
	socialgroupchecker "github.com/vibe-gaming/backend/internal/service/social_group_checker"
)

func newSocialGroupChecker(client *socialgroupchecker.Client, services *service.Services) SocialGroupChecker {
	return &socialGroupChecker{
		client:   client,
		services: services,
	}
}

type socialGroupChecker struct {
	client   *socialgroupchecker.Client
	services *service.Services
}

func (s *socialGroupChecker) CheckGroups(ctx context.Context, snils string, groups []socialgroupchecker.SocialGroup) (*socialgroupchecker.CheckResponse, error) {
	return s.client.CheckGroups(ctx, snils, groups)
}

// CheckAndUpdateUserGroups проверяет социальные группы пользователя и обновляет их статус в БД
func (s *socialGroupChecker) CheckAndUpdateUserGroups(ctx context.Context, userID uuid.UUID, snils string, groupTypes []string) error {
	// Получаем текущие данные пользователя
	user, err := s.services.Users.GetOneByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user by id failed: %w", err)
	}

	// Формируем список групп для проверки
	socialGroups := make([]socialgroupchecker.SocialGroup, 0, len(groupTypes))
	for _, gt := range groupTypes {
		socialGroups = append(socialGroups, socialgroupchecker.SocialGroup(gt))
	}

	// Вызываем внешний API для проверки
	checkResp, err := s.client.CheckGroups(ctx, snils, socialGroups)
	if err != nil {
		// Если произошла ошибка при проверке, помечаем все группы как отклонённые с сообщением об ошибке
		now := time.Now()
		updatedGroups := make(domain.UserGroupList, 0, len(user.GroupType))

		for _, userGroup := range user.GroupType {
			// Проверяем, входит ли эта группа в список на проверку
			shouldUpdate := false
			for _, gt := range groupTypes {
				if string(userGroup.Type) == gt {
					shouldUpdate = true
					break
				}
			}

			if shouldUpdate && userGroup.Status == domain.VerificationStatusPending {
				userGroup.Status = domain.VerificationStatusRejected
				userGroup.RejectedAt = &now
				userGroup.ErrorMessage = fmt.Sprintf("Ошибка проверки: %v", err)
			}
			updatedGroups = append(updatedGroups, userGroup)
		}

		if err := s.services.Users.UpdateUserGroups(ctx, userID, updatedGroups); err != nil {
			return fmt.Errorf("update user groups with error status failed: %w", err)
		}
		return fmt.Errorf("check groups failed: %w", err)
	}

	// Обновляем статус групп на основе ответа
	now := time.Now()
	updatedGroups := make(domain.UserGroupList, 0, len(user.GroupType))

	for _, userGroup := range user.GroupType {
		// Ищем результат проверки для этой группы
		for _, result := range checkResp.Results {
			if string(userGroup.Type) == string(result.Group) {
				if result.Status == socialgroupchecker.StatusConfirmed {
					userGroup.Status = domain.VerificationStatusVerified
					userGroup.VerifiedAt = &now
					// Установим срок действия через год
					expiresAt := now.AddDate(1, 0, 0)
					userGroup.ExpiresAt = &expiresAt
					userGroup.ErrorMessage = ""
				} else {
					userGroup.Status = domain.VerificationStatusRejected
					userGroup.RejectedAt = &now
					userGroup.ErrorMessage = ""
				}
				break
			}
		}
		updatedGroups = append(updatedGroups, userGroup)
	}

	// Сохраняем обновленные группы в БД
	if err := s.services.Users.UpdateUserGroups(ctx, userID, updatedGroups); err != nil {
		return fmt.Errorf("update user groups failed: %w", err)
	}

	return nil
}
