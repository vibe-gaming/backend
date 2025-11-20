package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/repository"
	logger "github.com/vibe-gaming/backend/pkg/logger"
	"github.com/vibe-gaming/backend/pkg/pdf"
	"go.uber.org/zap"
)

// BenefitFilters - псевдоним для удобства использования
type BenefitFilters = repository.BenefitFilters

// FilterStats - псевдоним для удобства использования
type FilterStats = repository.FilterStats

type BenefitService struct {
	benefitRepository      repository.BenefitRepository
	favoriteRepository     repository.FavoriteRepository
	usersRepository        repository.Users
	organizationRepository repository.OrganizationRepository
	gigachatClient         interface {
		EnhanceSearchQuery(ctx context.Context, query string) ([]string, error)
	}
}

func newBenefitService(
	benefitRepository repository.BenefitRepository,
	favoriteRepository repository.FavoriteRepository,
	userRepository repository.Users,
	organizationRepository repository.OrganizationRepository,
	gigachatClient interface {
		EnhanceSearchQuery(ctx context.Context, query string) ([]string, error)
	},
) *BenefitService {
	return &BenefitService{
		benefitRepository:      benefitRepository,
		favoriteRepository:     favoriteRepository,
		usersRepository:        userRepository,
		organizationRepository: organizationRepository,
		gigachatClient:         gigachatClient,
	}
}

func (s *BenefitService) GetAll(ctx context.Context, page, limit int, filters *BenefitFilters) ([]*domain.Benefit, int64, error) {
	// Валидация параметров пагинации
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Подготавливаем поисковый запрос для умного поиска
	if filters != nil && filters.Search != nil && *filters.Search != "" {
		originalQuery := *filters.Search
		logger.Info("Processing search query", zap.String("original_query", originalQuery))

		// Сначала пытаемся исправить распространенные опечатки
		correctedQuery := correctCommonTypos(originalQuery)
		if correctedQuery != originalQuery {
			logger.Info("Corrected typo in search query",
				zap.String("original", originalQuery),
				zap.String("corrected", correctedQuery))
			filters.Search = &correctedQuery
		}

		if containsBooleanOperators(*filters.Search) {
			// Пользователь использует свои операторы - не трогаем запрос
			logger.Info("User provided boolean operators, skipping GigaChat enhancement")
			filters.SearchMode = "boolean"
		} else {
			// Проверяем, что GigaChat клиент доступен
			if s.gigachatClient == nil {
				logger.Info("GigaChat client is nil, using fallback search")
				processedQuery := addWildcardsToQuery(*filters.Search)
				filters.Search = &processedQuery
				filters.SearchMode = "boolean"
			} else {
				// Используем GigaChat для улучшения поискового запроса
				logger.Info("Calling GigaChat to enhance search query")
				enhancedTerms, err := s.gigachatClient.EnhanceSearchQuery(ctx, *filters.Search)
				if err != nil {
					// Если GigaChat недоступен, используем обычный поиск
					logger.Error("GigaChat enhancement failed, using fallback search", zap.Error(err))
					processedQuery := addWildcardsToQuery(*filters.Search)
					filters.Search = &processedQuery
					filters.SearchMode = "boolean"
				} else {
					// Формируем Boolean запрос из расширенных терминов
					// Используем ИЛИ между терминами для максимального охвата
					logger.Info("GigaChat enhancement successful", zap.Strings("enhanced_terms", enhancedTerms))
					booleanQuery := buildBooleanQuery(enhancedTerms)
					logger.Info("Built boolean query", zap.String("query", booleanQuery))
					filters.Search = &booleanQuery
					filters.SearchMode = "boolean"
				}
			}
		}
	}

	offset := (page - 1) * limit

	benefits, err := s.benefitRepository.GetAll(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.benefitRepository.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return benefits, total, nil
}

// containsBooleanOperators проверяет, содержит ли поисковый запрос операторы Boolean режима
func containsBooleanOperators(query string) bool {
	// Boolean операторы MySQL Full-Text Search: +, -, *, ~, ", (, )
	booleanChars := []string{"+", "-", "*", "~", "\"", "(", ")"}
	for _, char := range booleanChars {
		if strings.Contains(query, char) {
			return true
		}
	}
	return false
}

// addWildcardsToQuery добавляет wildcard (*) к каждому слову для поиска по частичному совпадению
func addWildcardsToQuery(query string) string {
	// Убираем лишние пробелы
	query = strings.TrimSpace(query)
	if query == "" {
		return query
	}

	// Разбиваем на слова
	words := strings.Fields(query)

	// Добавляем * к каждому слову (если его там еще нет)
	processedWords := make([]string, 0, len(words))
	for _, word := range words {
		if !strings.HasSuffix(word, "*") {
			word = word + "*"
		}
		processedWords = append(processedWords, word)
	}

	// Собираем обратно
	return strings.Join(processedWords, " ")
}

// buildBooleanQuery создает Boolean запрос из списка терминов
// Используется для поиска по нескольким вариантам слов (с ошибками, морфологией, синонимами)
func buildBooleanQuery(terms []string) string {
	if len(terms) == 0 {
		return ""
	}

	// Добавляем wildcard к каждому термину и объединяем через OR (пробел в Boolean mode)
	processedTerms := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		// Для каждого термина добавляем wildcard для поиска по префиксу
		// Например: "студент*" найдет "студент", "студенты", "студентам" и т.д.
		if !strings.HasSuffix(term, "*") {
			term = term + "*"
		}
		processedTerms = append(processedTerms, term)
	}

	// В Boolean mode пробел между терминами означает OR
	// Это позволит найти документы, содержащие хотя бы один из терминов
	return strings.Join(processedTerms, " ")
}

// correctCommonTypos исправляет распространенные опечатки в поисковых запросах
// Это fallback на случай, если GigaChat недоступен или не распознал опечатку
func correctCommonTypos(query string) string {
	// Словарь распространенных опечаток и их исправлений
	// Ключ - опечатка (в нижнем регистре), значение - правильное написание
	typoMap := map[string]string{
		// Аптека
		"оптека": "аптека",
		"аптеко": "аптека",
		"оптеки": "аптека",
		"аптика": "аптека",

		// Пенсионер
		"пенсионир": "пенсионер",
		"пинсионер": "пенсионер",
		"пенсеонер": "пенсионер",

		// Инвалид
		"енвалид": "инвалид",
		"инвольд": "инвалид",
		"инволид": "инвалид",

		// Студент
		"стутент": "студент",
		"студэнт": "студент",

		// Транспорт
		"тронспорт":  "транспорт",
		"трансппорт": "транспорт",
		"трансопрт":  "транспорт",

		// Медицина
		"медецина": "медицина",
		"медицына": "медицина",
		"мидицина": "медицина",

		// Лекарство
		"ликарство": "лекарство",
		"лекорство": "лекарство",
		"лекарстов": "лекарство",

		// Скидка
		"скитка": "скидка",
		"сктдка": "скидка",
		"скдка":  "скидка",
	}

	// Разбиваем запрос на слова
	words := strings.Fields(query)
	correctedWords := make([]string, 0, len(words))

	for _, word := range words {
		// Проверяем каждое слово на опечатки
		lowerWord := strings.ToLower(word)
		if correction, exists := typoMap[lowerWord]; exists {
			// Сохраняем регистр первой буквы
			if len(word) > 0 {
				// Получаем первую руну слова
				firstRune := []rune(word)[0]
				// Проверяем, является ли она заглавной
				if firstRune >= 'А' && firstRune <= 'Я' {
					// Первая буква заглавная - делаем заглавную в исправлении
					correctionRunes := []rune(correction)
					if len(correctionRunes) > 0 {
						correctionRunes[0] = []rune(strings.ToUpper(string(correctionRunes[0])))[0]
						correction = string(correctionRunes)
					}
				}
			}
			correctedWords = append(correctedWords, correction)
		} else {
			// Слово не найдено в словаре - оставляем как есть
			correctedWords = append(correctedWords, word)
		}
	}

	return strings.Join(correctedWords, " ")
}

func (s *BenefitService) GetByID(ctx context.Context, id string, userID *uuid.UUID) (*domain.Benefit, error) {
	var userIDStr *string
	if userID != nil {
		userIDStrVal := userID.String()
		userIDStr = &userIDStrVal
	}

	benefit, err := s.benefitRepository.GetByID(ctx, id, userIDStr)
	if err != nil {
		return nil, err
	}

	if benefit.OrganizationID != nil {
		organization, err := s.organizationRepository.GetByID(ctx, benefit.OrganizationID.String())
		if err != nil {
			return nil, err
		}
		benefit.Organization = organization
	}

	return benefit, nil
}

func (s *BenefitService) IsFavorite(ctx context.Context, userID uuid.UUID, benefitID uuid.UUID) (bool, error) {
	favorite, err := s.favoriteRepository.GetByUserIDAndBenefitID(ctx, userID, benefitID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	// Проверяем, что запись не удалена (soft delete)
	return favorite.DeletedAt == nil, nil
}

func (s *BenefitService) MarkAsFavorite(ctx context.Context, userID uuid.UUID, benefitID uuid.UUID) error {
	// Проверяем, существует ли уже запись (включая удаленные)
	favorite, err := s.favoriteRepository.GetByUserIDAndBenefitID(ctx, userID, benefitID)
	if err != nil {
		// Если запись не найдена - создаем новую (добавляем в избранное)
		if errors.Is(err, domain.ErrNotFound) {
			return s.favoriteRepository.Create(ctx, &domain.Favorite{
				ID:        uuid.New(),
				UserID:    userID,
				BenefitID: benefitID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
		return err
	}

	now := time.Now()
	favorite.UpdatedAt = now

	// Toggle: если активна - удаляем, если удалена - восстанавливаем
	if favorite.DeletedAt == nil {
		// Запись активна → удаляем из избранного (soft delete)
		favorite.DeletedAt = &now
	} else {
		// Запись была удалена → восстанавливаем (добавляем обратно в избранное)
		favorite.DeletedAt = nil
	}

	return s.favoriteRepository.Update(ctx, favorite)
}

func (s *BenefitService) GetFilterStats(ctx context.Context, filters *BenefitFilters) (*FilterStats, error) {
	// Подготавливаем поисковый запрос для умного поиска (так же как в GetAll)
	if filters != nil && filters.Search != nil && *filters.Search != "" {
		originalQuery := *filters.Search

		// Сначала пытаемся исправить распространенные опечатки
		correctedQuery := correctCommonTypos(originalQuery)
		if correctedQuery != originalQuery {
			logger.Info("Corrected typo in search query for stats",
				zap.String("original", originalQuery),
				zap.String("corrected", correctedQuery))
			filters.Search = &correctedQuery
		}

		if containsBooleanOperators(*filters.Search) {
			// Пользователь использует свои операторы - не трогаем запрос
			filters.SearchMode = "boolean"
		} else {
			// Проверяем, что GigaChat клиент доступен
			if s.gigachatClient == nil {
				logger.Info("GigaChat client is nil in GetFilterStats, using fallback search")
				processedQuery := addWildcardsToQuery(*filters.Search)
				filters.Search = &processedQuery
				filters.SearchMode = "boolean"
			} else {
				// Используем GigaChat для улучшения поискового запроса
				enhancedTerms, err := s.gigachatClient.EnhanceSearchQuery(ctx, *filters.Search)
				if err != nil {
					// Если GigaChat недоступен, используем обычный поиск
					logger.Error("GigaChat enhancement failed in GetFilterStats", zap.Error(err))
					processedQuery := addWildcardsToQuery(*filters.Search)
					filters.Search = &processedQuery
					filters.SearchMode = "boolean"
				} else {
					// Формируем Boolean запрос из расширенных терминов
					booleanQuery := buildBooleanQuery(enhancedTerms)
					filters.Search = &booleanQuery
					filters.SearchMode = "boolean"
				}
			}
		}
	}

	return s.benefitRepository.GetFilterStats(ctx, filters)
}

func (s *BenefitService) GetUserBenefitsStats(ctx context.Context, userID uuid.UUID) (*repository.UserBenefitsStats, error) {

	user, err := s.usersRepository.GetOneByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Собираем подтвержденные группы пользователя
	targetGroups := []string{}
	for _, group := range user.GroupType {
		if group.Status == domain.VerificationStatusVerified {
			targetGroups = append(targetGroups, string(group.Type))
		}
	}

	logger.Info("Getting user benefits stats",
		zap.String("user_id", userID.String()),
		zap.Strings("target_groups", targetGroups))

	// Считаем доступные льготы для групп пользователя (OR логика)
	totalBenefits, err := s.benefitRepository.CountAvailableForUser(ctx, targetGroups)
	if err != nil {
		return nil, err
	}

	// Считаем избранные льготы
	favoritesCount, err := s.favoriteRepository.GetByUserCount(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &repository.UserBenefitsStats{
		TotalBenefits:  totalBenefits,
		TotalFavorites: favoritesCount,
	}, nil
}

func (s *BenefitService) GeneratePDF(ctx context.Context, benefit *domain.Benefit) ([]byte, error) {
	logger.Info("Generating PDF for benefit", zap.String("benefit_id", benefit.ID.String()))

	// Создаем генератор PDF
	generator := pdf.NewGenerator()

	// Генерируем PDF
	pdfBytes, err := generator.GenerateBenefitPDF(benefit)
	if err != nil {
		logger.Error("Failed to generate PDF", zap.Error(err), zap.String("benefit_id", benefit.ID.String()))
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	logger.Info("PDF generated successfully",
		zap.String("benefit_id", benefit.ID.String()),
		zap.Int("size_bytes", len(pdfBytes)))

	return pdfBytes, nil
}

func (s *BenefitService) GenerateBenefitsListPDF(ctx context.Context, benefits []*domain.Benefit, total int64, page int, limit int) ([]byte, error) {
	logger.Info("Generating PDF for benefits list",
		zap.Int("benefits_count", len(benefits)),
		zap.Int64("total", total),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// Создаем генератор PDF
	generator := pdf.NewGenerator()

	// Генерируем PDF
	pdfBytes, err := generator.GenerateBenefitsListPDF(benefits, int(total), page, limit)
	if err != nil {
		logger.Error("Failed to generate benefits list PDF",
			zap.Error(err),
			zap.Int("benefits_count", len(benefits)))
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	logger.Info("Benefits list PDF generated successfully",
		zap.Int("benefits_count", len(benefits)),
		zap.Int64("total", total),
		zap.Int("size_bytes", len(pdfBytes)))

	return pdfBytes, nil
}
