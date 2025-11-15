package v1

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/service"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func (h *Handler) initBenefits(api *gin.RouterGroup) {
	benefits := api.Group("/benefits")
	{
		benefits.GET("", h.getBenefitsList)
		benefits.GET("/:id", h.getBenefitByID)
	}
}

type benefitResponse struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ValidFrom    string   `json:"valid_from"`
	ValidTo      string   `json:"valid_to"`
	Type         string   `json:"type"`
	TargetGroups []string `json:"target_groups"`
	Longitude    *float64 `json:"longitude,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	CityID       *string  `json:"city_id,omitempty"`
	Region       []int    `json:"region"`
	Requirement  string   `json:"requirement"`
	HowToUse     *string  `json:"how_to_use,omitempty"`
	SourceURL    string   `json:"source_url"`
}

type benefitsListResponse struct {
	Benefits []benefitResponse `json:"benefits"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

// @Summary Get Benefits List
// @Tags Benefits
// @Description Получить список всех льгот с пагинацией, фильтрацией и умным поиском
// @Description
// @Description Поиск использует Full-Text Search MySQL с автоматическим поиском по частичному совпадению:
// @Description - Поиск "транс" найдет "транспорт", "транспортный" и т.д.
// @Description - Поиск "пенсион" найдет "пенсионер", "пенсионный" и т.д.
// @Description - Каждое слово ищется с начала (prefix matching)
// @Description
// @Description Можно использовать Boolean операторы для сложных запросов:
// @Description   + обязательное слово: "+пенсионер +транспорт"
// @Description   - исключить слово: "льгота -студент"
// @Description   * явный wildcard: "транс*"
// @Description   "" точная фраза: "общественный транспорт"
// @ModuleID getBenefitsList
// @Accept  json
// @Produce  json
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 10, максимум 100)"
// @Param region query int false "ID региона для фильтрации"
// @Param city_id query string false "UUID города для фильтрации"
// @Param type query string false "Тип льготы (federal, regional, commercial)"
// @Param target_groups query string false "Целевые группы через запятой (pensioners, disabled, students и т.д.)"
// @Param date_from query string false "Дата начала периода (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания периода (YYYY-MM-DD)"
// @Param search query string false "Поисковый запрос (автоматически ищет по частичному совпадению)"
// @Success 200 {object} benefitsListResponse
// @Failure 400 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /benefits [get]
func (h *Handler) getBenefitsList(c *gin.Context) {
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Собираем фильтры
	filters := &service.BenefitFilters{}

	if regionStr := c.Query("region"); regionStr != "" {
		if r, err := strconv.Atoi(regionStr); err == nil && r > 0 {
			filters.RegionID = &r
		}
	}

	if cityID := c.Query("city_id"); cityID != "" {
		filters.CityID = &cityID
	}

	if typeStr := c.Query("type"); typeStr != "" {
		filters.Type = &typeStr
	}

	if targetGroupsStr := c.Query("target_groups"); targetGroupsStr != "" {
		// Разделяем по запятой и убираем пробелы
		groups := strings.Split(targetGroupsStr, ",")
		for i, group := range groups {
			groups[i] = strings.TrimSpace(group)
		}
		filters.TargetGroups = groups
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters.DateFrom = &dateFrom
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		filters.DateTo = &dateTo
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	benefits, total, err := h.services.Benefits.GetAll(c.Request.Context(), page, limit, filters)
	if err != nil {
		logger.Error("failed to get benefits list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get benefits list"})
		return
	}

	response := benefitsListResponse{
		Benefits: make([]benefitResponse, 0, len(benefits)),
		Total:    total,
		Page:     page,
		Limit:    limit,
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

		response.Benefits = append(response.Benefits, benefitResponse{
			ID:           benefit.ID.String(),
			Title:        benefit.Title,
			Description:  benefit.Description,
			ValidFrom:    benefit.ValidFrom.Format("2006-01-02"),
			ValidTo:      benefit.ValidTo.Format("2006-01-02"),
			Type:         benefit.Type,
			TargetGroups: targetGroups,
			Longitude:    benefit.Longitude,
			Latitude:     benefit.Latitude,
			CityID:       cityID,
			Region:       benefit.Region,
			Requirement:  benefit.Requirement,
			HowToUse:     benefit.HowToUse,
			SourceURL:    benefit.SourceURL,
		})
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get Benefit By ID
// @Tags Benefits
// @Description Получить льготу по ID
// @ModuleID getBenefitByID
// @Accept  json
// @Produce  json
// @Param id path string true "Benefit ID"
// @Success 200 {object} benefitResponse
// @Failure 400 {object} ErrorStruct
// @Failure 404 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /benefits/{id} [get]
func (h *Handler) getBenefitByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("benefit id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "benefit id is required"})
		return
	}

	benefit, err := h.services.Benefits.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Error("benefit not found", zap.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "benefit not found"})
			return
		}
		logger.Error("failed to get benefit by id", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get benefit"})
		return
	}

	targetGroups := make([]string, 0, len(benefit.TargetGroupIDs))
	for _, tg := range benefit.TargetGroupIDs {
		targetGroups = append(targetGroups, string(tg))
	}

	var cityID *string
	if benefit.CityID != nil {
		cityIDStr := benefit.CityID.String()
		cityID = &cityIDStr
	}

	response := benefitResponse{
		ID:           benefit.ID.String(),
		Title:        benefit.Title,
		Description:  benefit.Description,
		ValidFrom:    benefit.ValidFrom.Format("2006-01-02"),
		ValidTo:      benefit.ValidTo.Format("2006-01-02"),
		Type:         benefit.Type,
		TargetGroups: targetGroups,
		Longitude:    benefit.Longitude,
		Latitude:     benefit.Latitude,
		CityID:       cityID,
		Region:       benefit.Region,
		Requirement:  benefit.Requirement,
		HowToUse:     benefit.HowToUse,
		SourceURL:    benefit.SourceURL,
	}

	c.JSON(http.StatusOK, response)
}
