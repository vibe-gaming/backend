package v1

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")

	users.GET("/pong", h.userIdentityMiddleware, h.pong)

	// auth routes
	users.GET("/auth/login", h.authLogin)
	users.GET("/auth/callback", h.authCallback)
}

// @Summary Pong
// @Tags Pong
// @Description Pong
// @ModuleID Pong
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400 {object} ErrorStruct
// @Failure 500
// @Security UserAuth
// @Router /users/pong [get]
func (h *Handler) pong(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

type userAuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken uuid.UUID `json:"refresh_token"`
}

// Хранилище для CSRF state токенов (в продакшене использовать Redis)
var stateStore = make(map[string]bool)

// @Summary OAuth Login
// @Tags Auth
// @Description Перенаправление на ESIA для авторизации
// @ModuleID login
// @Accept  json
// @Produce  json
// @Success 302
// @Router /users/auth/login [get]
func (h *Handler) authLogin(c *gin.Context) {
	// Генерируем state для защиты от CSRF
	state := generateState()
	stateStore[state] = true

	// Сохраняем state в cookie для проверки при callback
	c.SetCookie("esia_state", state, 600, "/", "", false, true)

	// Создаем ESIA клиент
	esiaClient := esia.NewClient(h.config.ESIA)

	// Получаем URL авторизации
	authURL := esiaClient.GetAuthorizationURL(state)

	logger.Info("Redirecting to ESIA", zap.String("url", authURL))

	// Перенаправляем пользователя на ESIA
	c.Redirect(http.StatusFound, authURL)
}

// @Summary OAuth Callback
// @Tags Auth
// @Description Callback endpoint для Auth
// @ModuleID callback
// @Accept  json
// @Produce  json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} userAuthResponse
// @Failure 400 {object} ErrorStruct
// @Router /users/auth/callback [get]
func (h *Handler) authCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		logger.Error("code is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// Проверяем state для защиты от CSRF
	savedState, err := c.Cookie("esia_state")
	if err != nil || savedState != state {
		logger.Error("invalid state", zap.String("saved", savedState), zap.String("received", state))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	// Проверяем, что state существует в нашем хранилище
	if !stateStore[state] {
		logger.Error("state not found in store")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	// Удаляем использованный state
	delete(stateStore, state)
	c.SetCookie("esia_state", "", -1, "/", "", false, true)

	logger.Info("ESIA callback received", zap.String("code", code[:10]+"..."))

	// Выполняем авторизацию через сервис
	result, err := h.services.Users.Auth(c.Request.Context(), code, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		logger.Error("esia auth failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization failed"})
		return
	}

	response := userAuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}

	logger.Info("ESIA auth successful")

	c.JSON(http.StatusOK, response)
}

// generateState генерирует случайный state для OAuth
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
