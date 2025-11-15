package apiHttp

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/vibe-gaming/backend/docs"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/limiter"
	"github.com/vibe-gaming/backend/pkg/logger"
	"github.com/vibe-gaming/backend/pkg/validator"

	internalV1 "github.com/vibe-gaming/backend/internal/api/http/internal/v1"
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services     *service.Services
	tokenManager auth.TokenManager
	config       *config.Config
}

func NewHandlers(services *service.Services, tokenManager auth.TokenManager, cfg *config.Config) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
		config:       cfg,
	}
}

func (h *Handler) Init(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	validator.RegisterGinValidator()

	router.Use(
		ginzap.Ginzap(logger.Logger(), time.RFC3339, true),
		limiter.Limit(cfg.Limiter.RPS, cfg.Limiter.Burst, cfg.Limiter.TTL),
		corsMiddleware,
	)
	router.Use(ginzap.RecoveryWithZap(logger.Logger(), true))

	if cfg.HttpServer.SwaggerEnabled {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("internal")))
	}

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	internalHandlersV1 := internalV1.NewHandler(h.services, h.tokenManager, h.config)
	api := router.Group("/api")
	internalHandlersV1.Init(api)
}
