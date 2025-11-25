package v1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/service"
	"github.com/vibe-gaming/backend/internal/service/gigachat"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func (h *Handler) initChatRoutes(api *gin.RouterGroup) {
	chat := api.Group("/chat")

	chat.POST("/start", h.userIdentityMiddleware, h.startChat)
	chat.POST("/message", h.userIdentityMiddleware, h.sendMessage)
}

type startChatResponse struct {
	Message string `json:"message"`
}

// @Summary Start Chat
// @Tags Chat
// @Description Start chat with user
// @ModuleID startChat
// @Accept json
// @Produce json
// @Success 200 {object} startChatResponse
// @Failure 400 {object} ErrorStruct
// @Security UserAuth
// @Router /chat/start [post]
func (h *Handler) startChat(c *gin.Context) {
	userId, err := h.getUserUUID(c)
	if err != nil {
		logger.Error("failed to get user UUID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user UUID"})
		return
	}

	user, err := h.services.Users.GetOneByID(c.Request.Context(), userId)
	if err != nil {
		logger.Error("failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	groupTypes := make([]string, len(user.GroupType))
	for i, groupType := range user.GroupType {
		groupTypes[i] = groupType.Type.StringLocale()
	}

	var message string
	if len(groupTypes) > 0 {
		message = "Здравствуйте, " + user.FirstName.String + "! Рад вас видеть. Я помогу вам найти подходящие льготы. Для вас доступны льготы по следующим категориям: " + strings.Join(groupTypes, ", ") + ". Чем могу помочь?"
	} else {
		message = "Здравствуйте, " + user.FirstName.String + "! Рад вас видеть. Я помогу вам найти подходящие льготы. Чем могу помочь?"
	}

	c.JSON(http.StatusOK, startChatResponse{
		Message: message,
	})
}

type sendMessageRequest struct {
	Message string `json:"message"`
}

type sendMessageResponse struct {
	Message  string            `json:"message"`
	Benefits []benefitResponse `json:"benefits,omitempty"`
}

// @Summary Send Message
// @Tags Chat
// @Description Send message to chat bot
// @ModuleID sendMessage
// @Accept json
// @Produce json
// @Param request body sendMessageRequest true "Message request"
// @Success 200 {object} sendMessageResponse
// @Failure 400 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Security UserAuth
// @Router /chat/message [post]
func (h *Handler) sendMessage(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	userId, err := h.getUserUUID(c)
	if err != nil {
		logger.Error("failed to get user UUID", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user UUID"})
		return
	}

	// Определяем намерение пользователя через Gigachat
	isBenefitRequest, searchQuery := h.detectBenefitIntent(c.Request.Context(), req.Message)
	logger.Info("detected benefit intent",
		zap.Bool("is_benefit_request", isBenefitRequest),
		zap.String("search_query", searchQuery),
		zap.String("original_message", req.Message))

	// Если searchQuery пустой, но это запрос льгот, пытаемся извлечь ключевые слова из сообщения
	if isBenefitRequest && searchQuery == "" {
		// Пытаемся извлечь ключевые слова из запроса
		searchQuery = h.extractKeywordsFromMessage(req.Message)
		logger.Info("extracted keywords from message",
			zap.String("extracted_query", searchQuery),
			zap.String("original_message", req.Message))
		if searchQuery == "" {
			// Если не удалось извлечь, используем оригинальное сообщение
			searchQuery = req.Message
		}
	}

	if isBenefitRequest && searchQuery != "" {
		// Пользователь просит льготы - используем тот же флоу что и getBenefits
		logger.Info("searching benefits", zap.String("search_query", searchQuery))
		benefits, err := h.searchBenefits(c.Request.Context(), userId, searchQuery)
		if err != nil {
			logger.Error("failed to search benefits", zap.Error(err), zap.String("search_query", searchQuery))
			// Если поиск не удался, отправляем в обычный чат
			response, err := h.sendToGigachat(c.Request.Context(), req.Message)
			if err != nil {
				logger.Error("failed to send to gigachat", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process message"})
				return
			}
			c.JSON(http.StatusOK, sendMessageResponse{
				Message:  response,
				Benefits: []benefitResponse{},
			})
			return
		}

		logger.Info("benefits found", zap.Int("count", len(benefits)), zap.String("search_query", searchQuery))

		// Формируем ответ с найденными льготами
		benefitResponses := h.mapBenefitsToResponse(c.Request.Context(), benefits, &userId)
		message := h.formatBenefitsResponse(len(benefits))

		c.JSON(http.StatusOK, sendMessageResponse{
			Message:  message,
			Benefits: benefitResponses,
		})
		return
	}

	// Пользователь не просит льготы - отправляем в обычный чат
	response, err := h.sendToGigachat(c.Request.Context(), req.Message)
	if err != nil {
		logger.Error("failed to send to gigachat", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process message"})
		return
	}

	c.JSON(http.StatusOK, sendMessageResponse{
		Message:  response,
		Benefits: []benefitResponse{},
	})
}

// detectBenefitIntent определяет, просит ли пользователь льготы
func (h *Handler) detectBenefitIntent(ctx context.Context, message string) (bool, string) {
	if h.gigachatClient == nil {
		// Если Gigachat недоступен, используем простую эвристику
		return h.simpleBenefitDetection(message), message
	}

	prompt := fmt.Sprintf(`Проанализируй сообщение пользователя и определи, просит ли он найти льготы, скидки или меры поддержки.

Сообщение: "%s"

Ответь ТОЛЬКО в формате JSON:
{
  "is_benefit_request": true/false,
  "search_query": "поисковый запрос для льгот или пустая строка"
}

ВАЖНО: 
- Если пользователь просит найти льготы, скидки, меры поддержки - установи is_benefit_request в true
- В search_query извлеки ТОЛЬКО ключевые слова темы запроса, БЕЗ слов "льготы", "скидки", "найди", "найти", "для", "по"
- search_query должен быть коротким (1-3 слова) и отражать суть запроса

Примеры:
"Найди мне льготы для лекарств" -> {"is_benefit_request": true, "search_query": "лекарства"}
"Какие льготы есть для пенсионеров?" -> {"is_benefit_request": true, "search_query": "пенсионеры"}
"Найди льготы для здоровья" -> {"is_benefit_request": true, "search_query": "здоровье"}
"Льготы на проезд" -> {"is_benefit_request": true, "search_query": "проезд"}
"Привет, как дела?" -> {"is_benefit_request": false, "search_query": ""}
"Расскажи о себе" -> {"is_benefit_request": false, "search_query": ""}`, message)

	reqBody := &gigachat.ChatRequest{
		Model: "GigaChat",
		Messages: []gigachat.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	chatResp, err := h.gigachatClient.Chat(reqBody)
	if err != nil {
		logger.Error("failed to detect benefit intent", zap.Error(err))
		return h.simpleBenefitDetection(message), message
	}

	if len(chatResp.Choices) == 0 {
		return h.simpleBenefitDetection(message), message
	}

	responseText := chatResp.Choices[0].Message.Content
	// Парсим JSON ответ
	// Упрощенный парсинг - ищем is_benefit_request и search_query
	if strings.Contains(responseText, `"is_benefit_request":true`) || strings.Contains(responseText, `"is_benefit_request": true`) {
		// Извлекаем search_query из JSON
		searchQuery := extractSearchQuery(responseText)
		return true, searchQuery
	}

	return false, ""
}

// simpleBenefitDetection простая эвристика для определения запроса льгот
func (h *Handler) simpleBenefitDetection(message string) bool {
	message = strings.ToLower(message)
	keywords := []string{"найди", "найти", "льгот", "скидк", "меры поддержки", "помощь", "какие льготы", "есть льготы"}
	for _, keyword := range keywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

// extractKeywordsFromMessage извлекает ключевые слова из сообщения пользователя
// Удаляет служебные слова и оставляет только значимые термины для поиска
func (h *Handler) extractKeywordsFromMessage(message string) string {
	message = strings.ToLower(message)

	// Удаляем служебные слова и фразы
	stopWords := []string{
		"найди", "найти", "мне", "для", "по", "льготы", "льгот", "скидки", "скидок",
		"меры поддержки", "помощь", "какие", "есть", "покажи", "дай", "дайте",
		"хочу", "нужны", "нужна", "нужно", "ищу", "ищем",
	}

	result := message
	for _, stopWord := range stopWords {
		result = strings.ReplaceAll(result, stopWord, " ")
	}

	// Очищаем от лишних пробелов
	result = strings.TrimSpace(result)
	result = strings.Join(strings.Fields(result), " ")

	return result
}

// extractSearchQuery извлекает поисковый запрос из JSON ответа
func extractSearchQuery(jsonText string) string {
	// Ищем "search_query": "значение"
	startIdx := strings.Index(jsonText, `"search_query"`)
	if startIdx == -1 {
		return ""
	}

	// Ищем двоеточие после "search_query"
	colonIdx := strings.Index(jsonText[startIdx:], `:`)
	if colonIdx == -1 {
		return ""
	}
	colonIdx = startIdx + colonIdx + 1

	// Пропускаем пробелы после двоеточия
	for colonIdx < len(jsonText) && (jsonText[colonIdx] == ' ' || jsonText[colonIdx] == '\t') {
		colonIdx++
	}

	// Ищем начало значения (открывающая кавычка)
	if colonIdx >= len(jsonText) || jsonText[colonIdx] != '"' {
		return ""
	}
	valueStart := colonIdx + 1

	// Ищем конец значения (закрывающая кавычка)
	valueEnd := strings.Index(jsonText[valueStart:], `"`)
	if valueEnd == -1 {
		return ""
	}

	return jsonText[valueStart : valueStart+valueEnd]
}

// searchBenefits ищет льготы используя тот же флоу что и getBenefits
func (h *Handler) searchBenefits(ctx context.Context, userId uuid.UUID, searchQuery string) ([]*domain.Benefit, error) {
	// Используем тот же флоу что и в getBenefits
	filters := &service.BenefitFilters{
		Search: &searchQuery,
	}

	// Получаем пользователя для фильтрации по группам
	user, err := h.services.Users.GetOneByID(ctx, userId)
	if err == nil && len(user.GroupType) > 0 {
		// Применяем фильтр по группам пользователя
		verifiedGroups := []string{}
		for _, group := range user.GroupType {
			if group.Status == domain.VerificationStatusVerified {
				verifiedGroups = append(verifiedGroups, string(group.Type))
			}
		}
		if len(verifiedGroups) > 0 {
			filterByUserGroups := true
			filters.FilterByUserGroups = &filterByUserGroups
			filters.UserGroupTypes = verifiedGroups
		}
		userIDStr := userId.String()
		filters.UserID = &userIDStr
	}

	// Используем сервис для поиска (там уже есть улучшение через Gigachat)
	benefits, _, err := h.services.Benefits.GetAll(ctx, 1, 10, filters)
	if err != nil {
		return nil, err
	}

	return benefits, nil
}

// sendToGigachat отправляет сообщение в обычный чат Gigachat
func (h *Handler) sendToGigachat(ctx context.Context, message string) (string, error) {
	if h.gigachatClient == nil {
		return "Извините, сервис временно недоступен.", nil
	}

	systemPrompt := `Ты помощник для поиска социальных льгот, коммерческих скидок и мер поддержки. 
Отвечай дружелюбно и по делу. Если пользователь спрашивает о льготах, предложи ему использовать поиск льгот.`

	reqBody := &gigachat.ChatRequest{
		Model: "GigaChat",
		Messages: []gigachat.Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: message,
			},
		},
	}

	chatResp, err := h.gigachatClient.Chat(reqBody)
	if err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "Извините, не удалось обработать ваш запрос.", nil
	}

	return chatResp.Choices[0].Message.Content, nil
}

// formatBenefitsResponse форматирует сообщение о найденных льготах
func (h *Handler) formatBenefitsResponse(count int) string {
	if count == 0 {
		return "К сожалению, я не нашел подходящих льгот по вашему запросу. Попробуйте изменить параметры поиска."
	}
	return fmt.Sprintf("Вот льготы, которые я нашел для вас (найдено: %d):", count)
}

// mapBenefitsToResponse преобразует доменные льготы в response формат
func (h *Handler) mapBenefitsToResponse(ctx context.Context, benefits []*domain.Benefit, userId *uuid.UUID) []benefitResponse {
	result := make([]benefitResponse, 0, len(benefits))

	var userID *uuid.UUID
	if userId != nil {
		userID = userId
	}

	for _, benefit := range benefits {
		targetGroups := make([]string, 0, len(benefit.TargetGroupIDs))
		for _, tg := range benefit.TargetGroupIDs {
			targetGroups = append(targetGroups, string(tg))
		}

		var cityID *string
		if benefit.CityID != nil {
			cityIDStr := benefit.CityID.String()
			cityID = &cityIDStr
		}

		var category *string
		if benefit.Category != nil {
			categoryStr := string(*benefit.Category)
			category = &categoryStr
		}

		tags := make([]string, 0, len(benefit.Tags))
		for _, tag := range benefit.Tags {
			tags = append(tags, string(tag))
		}

		var organization *organizationResponse
		if benefit.Organization != nil {
			organization = &organizationResponse{
				ID:          benefit.Organization.ID.String(),
				Name:        benefit.Organization.Name,
				Description: benefit.Organization.Description,
			}
			for i := range benefit.Organization.Buildings {
				building := &benefit.Organization.Buildings[i]
				organization.Buildings = append(organization.Buildings, organizationBuildingResponse{
					ID:          building.ID.String(),
					Address:     building.Address,
					Latitude:    building.Latitude,
					Longitude:   building.Longitude,
					PhoneNumber: building.PhoneNumber,
					GisDeeplink: building.GetGisDeeplink(),
					StartTime:   building.StartTime.Format("2006-01-02T15:04:05Z07:00"),
					EndTime:     building.EndTime.Format("2006-01-02T15:04:05Z07:00"),
					IsOpen:      building.IsOpen,
					Tags:        tags,
					Type:        building.Type,
				})
			}
		}

		// Проверяем, является ли льгота избранной
		isFavorite := false
		if userID != nil {
			favorite, err := h.services.Benefits.IsFavorite(ctx, *userID, benefit.ID)
			if err == nil {
				isFavorite = favorite
			}
		}

		result = append(result, benefitResponse{
			ID:           benefit.ID.String(),
			Title:        benefit.Title,
			Description:  benefit.Description,
			ValidFrom:    benefit.GetValidFrom(),
			ValidTo:      benefit.GetValidTo(),
			CreatedAt:    benefit.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    benefit.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Type:         string(benefit.Type),
			TargetGroups: targetGroups,
			Longitude:    benefit.Longitude,
			Latitude:     benefit.Latitude,
			CityID:       cityID,
			Region:       benefit.Region,
			Category:     category,
			Requirement:  benefit.Requirement,
			HowToUse:     benefit.HowToUse,
			SourceURL:    benefit.SourceURL,
			Tags:         tags,
			Views:        benefit.Views,
			GisDeeplink:  benefit.GetGisDeeplink(),
			Organization: organization,
			Favorite:     isFavorite,
		})
	}

	return result
}
