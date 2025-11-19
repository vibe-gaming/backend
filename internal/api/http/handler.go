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
		corsMiddleware([]string{"http://localhost:3000", "https://localhost:3000", "https://lgoty.netlify.app", "https://frontend-one-lovat-13.vercel.app"}),
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

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	internalHandlersV1 := internalV1.NewHandler(h.services, h.tokenManager, h.config, h.esiaClient, h.gigachatClient)
	api := router.Group("/api")
	internalHandlersV1.Init(api)
}
