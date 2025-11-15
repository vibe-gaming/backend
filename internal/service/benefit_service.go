package service

import (
	"context"
	"strings"

	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/repository"
)

// BenefitFilters - псевдоним для удобства использования
type BenefitFilters = repository.BenefitFilters

type BenefitService struct {
	benefitRepository repository.BenefitRepository
}

func newBenefitService(benefitRepository repository.BenefitRepository) *BenefitService {
	return &BenefitService{
		benefitRepository: benefitRepository,
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

	// Подготавливаем поисковый запрос для частичного поиска
	if filters != nil && filters.Search != nil && *filters.Search != "" {
		if containsBooleanOperators(*filters.Search) {
			// Пользователь использует свои операторы - не трогаем запрос
			filters.SearchMode = "boolean"
		} else {
			// Для обычного запроса добавляем * к каждому слову для prefix matching
			processedQuery := addWildcardsToQuery(*filters.Search)
			filters.Search = &processedQuery
			filters.SearchMode = "boolean"
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

func (s *BenefitService) GetByID(ctx context.Context, id string) (*domain.Benefit, error) {
	return s.benefitRepository.GetByID(ctx, id)
}
