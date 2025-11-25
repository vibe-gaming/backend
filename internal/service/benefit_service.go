package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
					// Используем оригинальный запрос как основу для обязательных терминов
					logger.Info("GigaChat enhancement successful", zap.Strings("enhanced_terms", enhancedTerms), zap.Int("enhanced_count", len(enhancedTerms)))
					originalQuery := *filters.Search
					booleanQuery := buildBooleanQuery(originalQuery, enhancedTerms)
					logger.Info("Built boolean query", zap.String("query", booleanQuery), zap.String("original_query", originalQuery))
					filters.Search = &booleanQuery
					filters.SearchMode = "boolean"
				}
			}
		}
	}

	offset := (page - 1) * limit

	// Если это запрос из чата, включен фильтр по группам пользователя и есть поисковый запрос,
	// проверяем, указывает ли запрос на другую целевую группу
	// Если да - ищем без фильтра по группам, если нет - используем фильтр строго
	// Для обычного списка льгот фильтры работают строго
	var originalFilterByUserGroups *bool
	var originalUserGroupTypes []string
	if filters != nil {
		originalFilterByUserGroups = filters.FilterByUserGroups
		originalUserGroupTypes = filters.UserGroupTypes
	}

	if filters != nil && filters.IsChatRequest && filters.FilterByUserGroups != nil && *filters.FilterByUserGroups && filters.Search != nil && *filters.Search != "" {
		// Проверяем, указывает ли запрос на другую целевую группу
		searchQueryLower := strings.ToLower(*filters.Search)
		userGroups := filters.UserGroupTypes
		if userGroups == nil {
			userGroups = []string{}
		}
		queryIndicatesOtherGroup := checkIfQueryIndicatesOtherGroup(searchQueryLower, userGroups)

		if queryIndicatesOtherGroup {
			// Запрос явно указывает на другую группу - ищем без фильтра по группам
			logger.Info("Query indicates other group in chat, searching without group filter",
				zap.String("search_query", *filters.Search),
				zap.Strings("user_groups", userGroups))

			// Отключаем фильтр по группам
			filters.FilterByUserGroups = nil
			filters.UserGroupTypes = nil

			// Выполняем поиск без фильтра
			benefits, err := s.benefitRepository.GetAll(ctx, limit, offset, filters)
			if err != nil {
				return nil, 0, err
			}

			total, err := s.benefitRepository.Count(ctx, filters)
			if err != nil {
				return nil, 0, err
			}

			logger.Info("Found results without group filter in chat",
				zap.Int64("total", total),
				zap.String("search_query", *filters.Search))

			return benefits, total, nil
		}

		// Запрос не указывает на другую группу - используем фильтр строго
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

	// Обычный поиск без специальной логики (для списка льгот фильтры работают строго)
	benefits, err := s.benefitRepository.GetAll(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.benefitRepository.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	// Восстанавливаем оригинальные значения фильтров (на случай если они были изменены)
	if originalFilterByUserGroups != nil {
		filters.FilterByUserGroups = originalFilterByUserGroups
		filters.UserGroupTypes = originalUserGroupTypes
	}

	return benefits, total, nil
}

// checkIfQueryIndicatesOtherGroup проверяет, указывает ли поисковый запрос на другую целевую группу,
// отличную от групп пользователя
func checkIfQueryIndicatesOtherGroup(searchQuery string, userGroups []string) bool {
	// Маппинг ключевых слов на группы
	groupKeywords := map[string][]string{
		"students":       {"студент", "студентам", "студента", "студенты", "студенческая", "студенческие", "студенческий", "вуз", "университет", "институт", "академия", "колледж", "обучение", "образование"},
		"pensioners":     {"пенсионер", "пенсионерам", "пенсионера", "пенсионеры", "пенсионная", "пенсионные"},
		"disabled":       {"инвалид", "инвалидам", "инвалида", "инвалиды", "инвалидность", "инвалидная", "инвалидные"},
		"veterans":       {"ветеран", "ветеранам", "ветерана", "ветераны", "ветеранская", "ветеранские"},
		"children":       {"ребенок", "детям", "детей", "дети", "детская", "детские", "детский", "школьник", "школьникам"},
		"young_families": {"молодая семья", "молодые семьи", "молодой семье", "молодых семей"},
		"large_families": {"многодетная семья", "многодетные семьи", "многодетной семье", "многодетных семей"},
		"low_income":     {"малоимущий", "малоимущим", "малоимущих", "малоимущие", "малообеспеченный", "малообеспеченным"},
	}

	// Проверяем, содержит ли запрос ключевые слова других групп
	for groupType, keywords := range groupKeywords {
		// Пропускаем группы пользователя
		isUserGroup := false
		for _, userGroup := range userGroups {
			if userGroup == groupType {
				isUserGroup = true
				break
			}
		}
		if isUserGroup {
			continue
		}

		// Проверяем наличие ключевых слов этой группы в запросе
		for _, keyword := range keywords {
			if strings.Contains(searchQuery, keyword) {
				logger.Info("Query indicates other group",
					zap.String("query", searchQuery),
					zap.String("indicated_group", groupType),
					zap.String("keyword", keyword),
					zap.Strings("user_groups", userGroups))
				return true
			}
		}
	}

	return false
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

// buildBooleanQuery создает Boolean запрос из оригинального запроса и расширенных терминов
// Используется для поиска по нескольким вариантам слов (с ошибками, морфологией, синонимами)
func buildBooleanQuery(originalQuery string, enhancedTerms []string) string {
	// Список служебных слов, которые не должны быть обязательными терминами
	// Эти слова часто встречаются в запросах, но не в описаниях льгот
	stopWords := map[string]bool{
		"льготы": true, "льгота": true, "льгот": true,
		"скидки": true, "скидка": true, "скидок": true,
		"найди": true, "найти": true, "найду": true, "найдет": true,
		"для": true, "по": true, "в": true, "на": true, "с": true,
		"мне": true, "меня": true,
		"какие": true, "какой": true, "какая": true,
		"есть": true, "имеются": true,
		"про": true, "о": true,
	}

	// Обрабатываем оригинальный запрос - разбиваем на слова
	originalWords := strings.Fields(strings.TrimSpace(originalQuery))

	// Фильтруем служебные слова и собираем значимые слова из оригинального запроса
	originalSignificantTerms := make([]string, 0, len(originalWords))
	for _, word := range originalWords {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		// Убираем знаки препинания в конце слова
		word = strings.TrimRight(word, ".,!?;:")
		if word == "" {
			continue
		}

		// Пропускаем служебные слова
		wordLower := strings.ToLower(word)
		if stopWords[wordLower] {
			continue
		}

		originalSignificantTerms = append(originalSignificantTerms, wordLower)
	}

	// Разделяем термины на основные (из оригинального запроса) и расширенные
	primaryTerms := make([]string, 0)
	enhancedTermsList := make([]string, 0)

	// Основные термины - из оригинального запроса
	for _, term := range originalSignificantTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		primaryTerms = append(primaryTerms, term)
	}

	// Расширенные термины - от GigaChat
	for _, term := range enhancedTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		termLower := strings.ToLower(term)

		// Пропускаем дубликаты
		isDuplicate := false
		for _, origTerm := range originalSignificantTerms {
			if origTerm == termLower {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}

		enhancedTermsList = append(enhancedTermsList, termLower)
	}

	// Обрабатываем основные термины: используем кавычки для фраз, wildcard для одиночных слов
	primaryProcessed := make([]string, 0)
	for _, term := range primaryTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		if strings.Contains(term, " ") {
			// Для фраз используем кавычки для точного поиска фразы
			// Оператор > повышает вес термина при сортировке
			primaryProcessed = append(primaryProcessed, ">"+`"`+term+`"`)
		} else {
			// Одиночное слово - добавляем wildcard
			if !strings.HasSuffix(term, "*") {
				term = term + "*"
			}
			// Оператор > повышает вес термина при сортировке
			primaryProcessed = append(primaryProcessed, ">"+term)
		}
	}

	// Обрабатываем расширенные термины: используем кавычки для фраз
	enhancedProcessed := make([]string, 0)
	for _, term := range enhancedTermsList {
		if strings.Contains(term, " ") {
			// Для фраз используем кавычки для точного поиска фразы в MySQL Boolean mode
			// Кавычки означают, что слова должны идти подряд в указанном порядке
			term = strings.TrimSpace(term)
			if term != "" {
				enhancedProcessed = append(enhancedProcessed, `"`+term+`"`)
			}
		} else {
			// Одиночное слово - добавляем wildcard
			if !strings.HasSuffix(term, "*") {
				term = term + "*"
			}
			enhancedProcessed = append(enhancedProcessed, term)
		}
	}

	// Сортируем термины для детерминированности (даже если они пришли в разном порядке)
	// Это обеспечит одинаковые SQL-запросы при одинаковых терминах
	sort.Strings(enhancedProcessed)

	// Ограничиваем количество расширенных терминов
	maxEnhanced := 10
	if len(enhancedProcessed) > maxEnhanced {
		enhancedProcessed = enhancedProcessed[:maxEnhanced]
	}

	// Формируем финальный запрос:
	// - Все термины опциональны (OR логика)
	// - Основные термины с повышенным весом (>) для лучшей сортировки - они будут выше в результатах
	// - Расширенные термины опциональны без дополнительного веса

	if len(primaryProcessed) == 0 && len(enhancedProcessed) == 0 {
		return ""
	}

	allTerms := make([]string, 0)

	// Добавляем основные термины с повышенным весом (они будут выше в сортировке)
	if len(primaryProcessed) > 0 {
		allTerms = append(allTerms, primaryProcessed...)
	}

	// Добавляем расширенные термины как опциональные
	if len(enhancedProcessed) > 0 {
		allTerms = append(allTerms, enhancedProcessed...)
	}

	// Все термины объединяются через пробел (OR логика)
	// Термины с оператором > будут иметь больший вес при сортировке
	return strings.Join(allTerms, " ")
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

	// Убеждаемся, что теги не nil перед обновлением
	if benefit.Tags == nil {
		benefit.Tags = domain.BenefitTagList{}
	}

	benefit.Views++
	err = s.benefitRepository.Update(ctx, benefit)
	if err != nil {
		return nil, err
	}

	return benefit, nil
}

// GetByIDWithoutIncrement получает льготу по ID без увеличения счетчика просмотров
func (s *BenefitService) GetByIDWithoutIncrement(ctx context.Context, id string, userID *uuid.UUID) (*domain.Benefit, error) {
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

	// Убеждаемся, что теги не nil
	if benefit.Tags == nil {
		benefit.Tags = domain.BenefitTagList{}
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
					originalQuery := *filters.Search
					booleanQuery := buildBooleanQuery(originalQuery, enhancedTerms)
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

func (s *BenefitService) Count(ctx context.Context, filters *BenefitFilters) (int64, error) {
	return s.benefitRepository.Count(ctx, filters)
}

func (s *BenefitService) GetBenefitTypesStats(ctx context.Context) (map[string]int64, error) {
	stats, err := s.benefitRepository.GetFilterStats(ctx, nil)
	if err != nil {
		return nil, err
	}
	return stats.Levels, nil
}

func (s *BenefitService) Update(ctx context.Context, benefit *domain.Benefit) error {
	benefit.UpdatedAt = time.Now()
	// Убеждаемся, что теги не nil
	if benefit.Tags == nil {
		benefit.Tags = domain.BenefitTagList{}
	}
	return s.benefitRepository.Update(ctx, benefit)
}

func (s *BenefitService) Delete(ctx context.Context, id string) error {
	return s.benefitRepository.Delete(ctx, id)
}

func (s *BenefitService) Create(ctx context.Context, benefit *domain.Benefit) error {
	if benefit.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generate benefit id failed: %w", err)
		}
		benefit.ID = newID
	}

	now := time.Now()
	if benefit.CreatedAt.IsZero() {
		benefit.CreatedAt = now
	}
	if benefit.UpdatedAt.IsZero() {
		benefit.UpdatedAt = now
	}

	return s.benefitRepository.Create(ctx, benefit)
}
