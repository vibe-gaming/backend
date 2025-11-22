package apiHttp

import (
	"io"
	"net/http"
	"os"
	"time"

	ginzap "github.com/gin-contrib/zap"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/vibe-gaming/backend/docs"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/limiter"
	"github.com/vibe-gaming/backend/pkg/logger"
	"github.com/vibe-gaming/backend/pkg/validator"
	"go.uber.org/zap"

	"github.com/vibe-gaming/backend/internal/api/http/admin"
	internalV1 "github.com/vibe-gaming/backend/internal/api/http/internal/v1"
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/internal/service"
	"github.com/vibe-gaming/backend/internal/service/gigachat"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services       *service.Services
	tokenManager   auth.TokenManager
	config         *config.Config
	esiaClient     *esia.Client
	gigachatClient *gigachat.Client
}

func NewHandlers(
	services *service.Services,
	tokenManager auth.TokenManager,
	cfg *config.Config,
	esiaClient *esia.Client,
	gigachatClient *gigachat.Client,
) *Handler {
	return &Handler{
		services:       services,
		tokenManager:   tokenManager,
		config:         cfg,
		esiaClient:     esiaClient,
		gigachatClient: gigachatClient,
	}
}

func (h *Handler) Init(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	validator.RegisterGinValidator()

	router.Use(
		ginzap.Ginzap(logger.Logger(), time.RFC3339, true),
		limiter.Limit(cfg.Limiter.RPS, cfg.Limiter.Burst, cfg.Limiter.TTL),
		// TODO: Get from config
		corsMiddleware([]string{"http://localhost:3000", "https://localhost:3000", "https://lgoty.netlify.app", "https://frontend-one-lovat-13.vercel.app", "https://localhost:3001", "https://frontend-two-steel-94.vercel.app", "https://frontend-production-9c0b.up.railway.app"}),
	)
	router.Use(ginzap.RecoveryWithZap(logger.Logger(), true))

	if cfg.HttpServer.SwaggerEnabled {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("internal")))
	}

	router.GET("/speech-test", func(c *gin.Context) {
		file, err := os.Open("./internal/api/http/speech_test.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to open file")
			return
		}
		defer file.Close()

		htmlContent, err := io.ReadAll(file)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read file")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", htmlContent)
	})

	// Инициализируем admin handler
	adminHandler := admin.NewHandler()

	// Админка на корневом уровне
	router.GET("/admin/", adminHandler.AdminPage)
	router.GET("/admin/stats", h.getAdminStats)
	router.GET("/admin/create-benefit", adminHandler.CreateBenefitPage)
	router.GET("/admin/benefits", adminHandler.BenefitsListPage)
	router.GET("/admin/benefits/:id", adminHandler.BenefitDetailPage)
	router.GET("/admin/organizations", adminHandler.OrganizationsListPage)
	router.GET("/admin/organizations/:id", adminHandler.OrganizationDetailPage)
	router.GET("/admin/create-organization", adminHandler.CreateOrganizationPage)

	h.initAdminRoutes(router)
	h.initAPI(router)

	return router
}

func (h *Handler) initAdminRoutes(router *gin.Engine) {
	api := router.Group("/api")
	adminGroup := api.Group("/admin")
	adminHandler := admin.NewHandler()
	adminGroup.GET("/", adminHandler.AdminPage)
	adminGroup.GET("/stats", h.getAdminStats)
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

func (h *Handler) initAPI(router *gin.Engine) {
	internalHandlersV1 := internalV1.NewHandler(h.services, h.tokenManager, h.config, h.esiaClient, h.gigachatClient)
	api := router.Group("/api")
	internalHandlersV1.Init(api)
}
