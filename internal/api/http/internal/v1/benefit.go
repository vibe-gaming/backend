package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/service"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func (h *Handler) initBenefits(api *gin.RouterGroup) {
	benefits := api.Group("/benefits")
	{
		benefits.GET("", h.optionalUserIdentityMiddleware, h.getBenefitsList)
		benefits.GET("/stats", h.optionalUserIdentityMiddleware, h.getBenefitsFilterStats)
		benefits.GET("/:id", h.getBenefitByID)
		benefits.POST("/:id/favorite", h.userIdentityMiddleware, h.markBenefitAsFavorite)
		benefits.GET("/user-stats", h.userIdentityMiddleware, h.getUserBenefitsStats)
		benefits.GET("/:id/pdfdownload", h.getBenefitPDFDownload)
	}
}

type benefitResponse struct {
	ID           string                `json:"id"`
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	ValidFrom    string                `json:"valid_from"`
	ValidTo      string                `json:"valid_to"`
	CreatedAt    string                `json:"created_at"`
	UpdatedAt    string                `json:"updated_at"`
	Type         string                `json:"type"`
	TargetGroups []string              `json:"target_groups"`
	Longitude    *float64              `json:"longitude,omitempty"`
	Latitude     *float64              `json:"latitude,omitempty"`
	CityID       *string               `json:"city_id,omitempty"`
	Region       []int                 `json:"region"`
	Category     *string               `json:"category,omitempty"`
	Requirement  string                `json:"requirement"`
	HowToUse     *string               `json:"how_to_use,omitempty"`
	SourceURL    string                `json:"source_url"`
	Tags         []string              `json:"tags"`
	Views        int                   `json:"views"`
	GisDeeplink  string                `json:"gis_deeplink,omitempty"`
	Organization *organizationResponse `json:"organization,omitempty"`
}

type organizationResponse struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Buildings   []organizationBuildingResponse `json:"buildings"`
}

type organizationBuildingResponse struct {
	ID          string   `json:"id"`
	Address     string   `json:"address"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	PhoneNumber string   `json:"phone_number"`
	GisDeeplink string   `json:"gis_deeplink"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	IsOpen      bool     `json:"is_open"`
	Tags        []string `json:"tags"`
	Type        string   `json:"type"`
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
// @Param target_groups query string false "Целевые группы через запятую (pensioners, disabled, students и т.д.)"
// @Param tags query string false "Теги через запятую (most_popular, new, hot, best, recommended, popular, top)"
// @Param categories query string false "Категории через запятую (medicine, transport, food, clothing, other)"
// @Param date_from query string false "Дата начала периода (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания периода (YYYY-MM-DD)"
// @Param search query string false "Поисковый запрос (автоматически ищет по частичному совпадению)"
// @Param sort_by query string false "Поле для сортировки (created_at, views, updated_at) - по умолчанию created_at"
// @Param order query string false "Направление сортировки (asc, desc) - по умолчанию desc"
// @Param favorites query boolean false "Показать только избранные льготы (работает только при авторизации, иначе игнорируется)"
// @Param filter_by_user_groups query boolean false "Фильтровать льготы по группам пользователя (работает только при авторизации)"
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

	if tagsStr := c.Query("tags"); tagsStr != "" {
		// Разделяем по запятой и убираем пробелы
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		filters.Tags = tags
	}

	if categoriesStr := c.Query("categories"); categoriesStr != "" {
		// Разделяем по запятой и убираем пробелы
		categories := strings.Split(categoriesStr, ",")
		for i, category := range categories {
			categories[i] = strings.TrimSpace(category)
		}
		filters.Categories = categories
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

	// Фильтр по избранным (только для авторизованных пользователей)
	if favoritesStr := c.Query("favorites"); favoritesStr == "true" {
		// Пытаемся получить userID из контекста (если пользователь авторизован)
		if userID, err := h.getUserUUID(c); err == nil {
			userIDStr := userID.String()
			filters.UserID = &userIDStr
			logger.Info("favorites filter applied", zap.String("user_id", userIDStr))
		} else {
			logger.Info("favorites filter requested but user not authenticated", zap.Error(err))
		}
	}

	// Фильтр по группам пользователя (только для авторизованных пользователей)
	if filterByUserGroupsStr := c.Query("filter_by_user_groups"); filterByUserGroupsStr == "true" {
		if userID, err := h.getUserUUID(c); err == nil {
			logger.Info("filter_by_user_groups=true received", zap.String("user_id", userID.String()))

			// Получаем пользователя из репозитория
			user, err := h.services.Users.GetOneByID(c.Request.Context(), userID)
			if err != nil {
				logger.Error("failed to get user for group filtering", zap.Error(err), zap.String("user_id", userID.String()))
			} else {
				logger.Info("user retrieved successfully",
					zap.String("user_id", userID.String()),
					zap.Int("total_groups", len(user.GroupType)))

				// Собираем подтвержденные группы пользователя
				verifiedGroups := []string{}
				for _, group := range user.GroupType {
					logger.Info("processing user group",
						zap.String("group_type", string(group.Type)),
						zap.String("status", string(group.Status)))

					if group.Status == domain.VerificationStatusVerified {
						verifiedGroups = append(verifiedGroups, string(group.Type))
					}
				}

				// Всегда применяем фильтр, даже если групп нет
				// Если групп нет - вернется пустой результат
				filterByUserGroups := true
				filters.FilterByUserGroups = &filterByUserGroups
				filters.UserGroupTypes = verifiedGroups

				if len(verifiedGroups) > 0 {
					logger.Info("user groups filter applied",
						zap.String("user_id", userID.String()),
						zap.Strings("verified_groups", verifiedGroups))
				} else {
					logger.Info("user has no verified groups, filter will return empty results",
						zap.String("user_id", userID.String()))
				}
			}
		} else {
			logger.Info("filter_by_user_groups requested but user not authenticated", zap.Error(err))
		}
	}

	// Параметры сортировки
	sortBy := c.Query("sort_by")
	if sortBy != "" {
		// Валидация допустимых значений
		switch sortBy {
		case "created_at", "views", "updated_at":
			filters.SortBy = sortBy
		default:
			filters.SortBy = "created_at"
		}
	} else {
		filters.SortBy = "created_at"
	}

	order := c.Query("order")
	if order != "" {
		// Валидация допустимых значений
		switch strings.ToLower(order) {
		case "asc", "desc":
			filters.Order = strings.ToLower(order)
		default:
			filters.Order = "desc"
		}
	} else {
		filters.Order = "desc"
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
				logger.Info("building coordinates",
					zap.String("id", building.ID.String()),
					zap.Float64("latitude", building.Latitude),
					zap.Float64("longitude", building.Longitude),
					zap.String("gis_deeplink", building.GetGisDeeplink()),
				)
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

		response.Benefits = append(response.Benefits, benefitResponse{
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
			logger.Info("building coordinates in getBenefitByID",
				zap.String("id", building.ID.String()),
				zap.Float64("latitude", building.Latitude),
				zap.Float64("longitude", building.Longitude),
				zap.String("gis_deeplink", building.GetGisDeeplink()),
			)

			var buildingTags []string

			for _, buildingTag := range building.Tags {
				buildingTags = append(buildingTags, string(buildingTag))
			}
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
				Tags:        buildingTags,
				Type:        building.Type,
			})
		}
	}
	response := benefitResponse{
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
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get Benefits Filter Statistics
// @Tags Benefits
// @Description Получить статистику по фильтрам - количество льгот по категориям и уровням
// @Description
// @Description Поддерживает те же параметры фильтрации что и GET /benefits (кроме category и type, так как мы их считаем)
// @Description Это позволяет показывать актуальные счетчики в форме фильтров при изменении других параметров
// @ModuleID getBenefitsFilterStats
// @Accept  json
// @Produce  json
// @Param region query int false "ID региона для фильтрации"
// @Param city_id query string false "UUID города для фильтрации"
// @Param target_groups query string false "Целевые группы через запятую"
// @Param tags query string false "Теги через запятую"
// @Param date_from query string false "Дата начала периода (YYYY-MM-DD)"
// @Param date_to query string false "Дата окончания периода (YYYY-MM-DD)"
// @Param search query string false "Поисковый запрос"
// @Param favorites query boolean false "Учитывать только избранные льготы (работает только при авторизации)"
// @Param filter_by_user_groups query boolean false "Фильтровать льготы по группам пользователя (работает только при авторизации)"
// @Success 200 {object} map[string]map[string]int64
// @Failure 500 {object} ErrorStruct
// @Router /benefits/stats [get]
func (h *Handler) getBenefitsFilterStats(c *gin.Context) {
	// Собираем фильтры (те же что и в getBenefitsList, но без категорий и типов)
	filters := &service.BenefitFilters{}

	if regionStr := c.Query("region"); regionStr != "" {
		if r, err := strconv.Atoi(regionStr); err == nil && r > 0 {
			filters.RegionID = &r
		}
	}

	if cityID := c.Query("city_id"); cityID != "" {
		filters.CityID = &cityID
	}

	if targetGroupsStr := c.Query("target_groups"); targetGroupsStr != "" {
		groups := strings.Split(targetGroupsStr, ",")
		for i, group := range groups {
			groups[i] = strings.TrimSpace(group)
		}
		filters.TargetGroups = groups
	}

	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		filters.Tags = tags
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

	// Фильтр по избранным (только для авторизованных пользователей)
	if favoritesStr := c.Query("favorites"); favoritesStr == "true" {
		if userID, err := h.getUserUUID(c); err == nil {
			userIDStr := userID.String()
			filters.UserID = &userIDStr
		}
	}

	// Фильтр по группам пользователя (только для авторизованных пользователей)
	if filterByUserGroupsStr := c.Query("filter_by_user_groups"); filterByUserGroupsStr == "true" {
		if userID, err := h.getUserUUID(c); err == nil {
			logger.Info("filter_by_user_groups=true received in stats", zap.String("user_id", userID.String()))

			// Получаем пользователя из репозитория
			user, err := h.services.Users.GetOneByID(c.Request.Context(), userID)
			if err != nil {
				logger.Error("failed to get user for group filtering in stats", zap.Error(err), zap.String("user_id", userID.String()))
			} else {
				logger.Info("user retrieved successfully in stats",
					zap.String("user_id", userID.String()),
					zap.Int("total_groups", len(user.GroupType)))

				// Собираем подтвержденные группы пользователя
				verifiedGroups := []string{}
				for _, group := range user.GroupType {
					logger.Info("processing user group in stats",
						zap.String("group_type", string(group.Type)),
						zap.String("status", string(group.Status)))

					if group.Status == domain.VerificationStatusVerified {
						verifiedGroups = append(verifiedGroups, string(group.Type))
					}
				}

				// Всегда применяем фильтр, даже если групп нет
				// Если групп нет - вернется пустой результат
				filterByUserGroups := true
				filters.FilterByUserGroups = &filterByUserGroups
				filters.UserGroupTypes = verifiedGroups

				if len(verifiedGroups) > 0 {
					logger.Info("user groups filter applied in stats",
						zap.String("user_id", userID.String()),
						zap.Strings("verified_groups", verifiedGroups))
				} else {
					logger.Info("user has no verified groups in stats, filter will return empty results",
						zap.String("user_id", userID.String()))
				}
			}
		} else {
			logger.Info("filter_by_user_groups requested in stats but user not authenticated", zap.Error(err))
		}
	}

	stats, err := h.services.Benefits.GetFilterStats(c.Request.Context(), filters)
	if err != nil {
		logger.Error("failed to get filter stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get filter stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Toggle Benefit Favorite Status
// @Tags Benefits
// @Description Переключить статус избранной льготы (toggle). Если льгота в избранном - удалит из избранного, если нет - добавит в избранное
// @ModuleID markBenefitAsFavorite
// @Accept  json
// @Produce  json
// @Param id path string true "Benefit ID (UUID)"
// @Security UserAuth
// @Success 200
// @Failure 400 {object} ErrorStruct
// @Failure 401 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /benefits/{id}/favorite [post]
func (h *Handler) markBenefitAsFavorite(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("benefit id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "benefit id is required"})
		return
	}

	userID, err := h.getUserUUID(c)
	if err != nil {
		logger.Error("failed to get user id", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user id"})
		return
	}

	err = h.services.Benefits.MarkAsFavorite(c.Request.Context(), userID, uuid.MustParse(id))
	if err != nil {
		logger.Error("failed to mark benefit as favorite", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark benefit as favorite"})
		return
	}

	c.Status(http.StatusOK)
}

// @Summary Get User Benefits Stats
// @Tags Benefits
// @Description Получить статистику по льготам пользователя
// @ModuleID getUserBenefitsStats
// @Accept  json
// @Produce  json
// @Security UserAuth
// @Success 200 {object} repository.UserBenefitsStats
// @Failure 500 {object} ErrorStruct
// @Router /benefits/user-stats [get]
func (h *Handler) getUserBenefitsStats(c *gin.Context) {
	userID, err := h.getUserUUID(c)
	if err != nil {
		logger.Error("failed to get user id", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user id"})
		return
	}

	stats, err := h.services.Benefits.GetUserBenefitsStats(c.Request.Context(), userID)
	if err != nil {
		logger.Error("failed to get user benefits stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user benefits stats"})

		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get Benefit PDF Download
// @Tags Benefits
// @Description Скачать льготу в формате PDF
// @ModuleID getBenefitPDFDownload
// @Accept  json
// @Produce  application/pdf
// @Param id path string true "Benefit ID (UUID)"
// @Success 200 {file} application/pdf
// @Failure 400 {object} ErrorStruct
// @Failure 404 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /benefits/{id}/pdfdownload [get]
func (h *Handler) getBenefitPDFDownload(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("benefit id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "benefit id is required"})
		return
	}

	// Получаем льготу
	benefit, err := h.services.Benefits.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Error("benefit not found for PDF", zap.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "benefit not found"})
			return
		}
		logger.Error("failed to get benefit for PDF", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get benefit"})
		return
	}

	// Генерируем PDF
	pdfBytes, err := h.services.Benefits.GeneratePDF(c.Request.Context(), benefit)
	if err != nil {
		logger.Error("failed to generate pdf", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pdf"})
		return
	}

	// Формируем имя файла
	filename := fmt.Sprintf("benefit_%s.pdf", benefit.ID.String())

	// Устанавливаем заголовки для скачивания файла
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Отправляем PDF
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
