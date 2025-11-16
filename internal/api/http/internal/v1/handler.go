package v1

import (
	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/internal/service"
	"github.com/vibe-gaming/backend/internal/service/gigachat"
	"github.com/vibe-gaming/backend/internal/service/yandexgpt"
	"github.com/vibe-gaming/backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

// @title Backend API
// @version 1.0
// @description Backend API

// @BasePath /api/v1

// @securityDefinitions.apikey AdminAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey UserAuth
// @in header
// @name Authorization

type Handler struct {
	services       *service.Services
	tokenManager   auth.TokenManager
	config         *config.Config
	esiaClient     *esia.Client
	gigachatClient *gigachat.Client
	yandexClient   *yandexgpt.Client
}

func NewHandler(
	services *service.Services,
	tokenManager auth.TokenManager,
	config *config.Config,
	esiaClient *esia.Client,
	gigachatClient *gigachat.Client,
	yandexClient *yandexgpt.Client,
) *Handler {
	return &Handler{
		services:       services,
		tokenManager:   tokenManager,
		config:         config,
		esiaClient:     esiaClient,
		gigachatClient: gigachatClient,
		yandexClient:   yandexClient,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("v1")

	h.initUsersRoutes(v1)
	h.initBenefits(v1)
	h.initCitiesRoutes(v1)
	h.initSpeechRoutes(v1)
}
