package v1

import (
	"io"
	"net/http"
	"os"

	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initAdminRoutes(api *gin.RouterGroup) {
	admin := api.Group("/admin")
	admin.GET("/", h.adminPage)
	admin.GET("/stats", h.getAdminStats)
}

func (h *Handler) adminPage(c *gin.Context) {
	file, err := os.Open("./internal/api/http/admin.html")
	if err != nil {
		logger.Error("failed to open admin.html", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to open admin page")
		return
	}
	defer file.Close()

	htmlContent, err := io.ReadAll(file)
	if err != nil {
		logger.Error("failed to read admin.html", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to read admin page")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", htmlContent)
}

type adminStatsResponse struct {
	TotalUsers     int64            `json:"total_users"`
	TotalBenefits  int64            `json:"total_benefits"`
	TotalCities    int64            `json:"total_cities"`
	TotalFavorites int64            `json:"total_favorites"`
	UserGroups     map[string]int64 `json:"user_groups"`
	BenefitTypes   map[string]int64 `json:"benefit_types"`
}

func (h *Handler) getAdminStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем статистику пользователей
	totalUsers, err := h.services.Users.Count(ctx)
	if err != nil {
		logger.Error("failed to get users count", zap.Error(err))
		totalUsers = 0
	}

	// Получаем статистику льгот
	totalBenefits, err := h.services.Benefits.Count(ctx, nil)
	if err != nil {
		logger.Error("failed to get benefits count", zap.Error(err))
		totalBenefits = 0
	}

	// Получаем статистику городов
	totalCities, err := h.services.Cities.Count(ctx)
	if err != nil {
		logger.Error("failed to get cities count", zap.Error(err))
		totalCities = 0
	}

	// Получаем статистику избранных
	totalFavorites, err := h.services.Favorites.GetTotalCount(ctx)
	if err != nil {
		logger.Error("failed to get favorites count", zap.Error(err))
		totalFavorites = 0
	}

	// Получаем статистику по группам пользователей
	userGroups, err := h.services.Users.GetUserGroupsStats(ctx)
	if err != nil {
		logger.Error("failed to get user groups stats", zap.Error(err))
		userGroups = make(map[string]int64)
	}

	// Получаем статистику по типам льгот
	benefitTypes, err := h.services.Benefits.GetBenefitTypesStats(ctx)
	if err != nil {
		logger.Error("failed to get benefit types stats", zap.Error(err))
		benefitTypes = make(map[string]int64)
	}

	response := adminStatsResponse{
		TotalUsers:     totalUsers,
		TotalBenefits:  totalBenefits,
		TotalCities:    totalCities,
		TotalFavorites: totalFavorites,
		UserGroups:     userGroups,
		BenefitTypes:   benefitTypes,
	}

	c.JSON(http.StatusOK, response)
}
