package v1

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// Хранилище для CSRF state токенов (в продакшене использовать Redis)
var stateStore = make(map[string]bool)

// Хранилище для authorization codes с привязкой к state
type authCodeData struct {
	Code      string
	State     string
	CreatedAt time.Time
}

var codeStore = make(map[string]*authCodeData)
var codeStoreMutex sync.RWMutex

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")

	users.GET("/pong", h.userIdentityMiddleware, h.pong)

	// auth routes
	users.GET("/auth/login", h.authLogin)
	users.GET("/auth/callback", h.authCallback)
	users.POST("/auth/token", h.exchangeToken)
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
// @Description Callback endpoint для Auth - получает code от ESIA и редиректит на фронтенд
// @ModuleID callback
// @Accept  json
// @Produce  json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 302 "Redirect to frontend with code"
// @Failure 400 {object} ErrorStruct
// @Router /users/auth/callback [get]
func (h *Handler) authCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		logger.Error("code is empty")
		frontendURL := fmt.Sprintf("%s/auth/error?error=no_code", h.config.FrontendURL)
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	// Проверяем state для защиты от CSRF
	savedState, err := c.Cookie("esia_state")
	if err != nil || savedState != state {
		logger.Error("invalid state", zap.String("saved", savedState), zap.String("received", state))
		frontendURL := fmt.Sprintf("%s/auth/error?error=invalid_state", h.config.FrontendURL)
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	// Проверяем, что state существует в нашем хранилище
	if !stateStore[state] {
		logger.Error("state not found in store")
		frontendURL := fmt.Sprintf("%s/auth/error?error=state_not_found", h.config.FrontendURL)
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	// Сохраняем code для последующего обмена на токены
	codeStoreMutex.Lock()
	codeStore[code] = &authCodeData{
		Code:      code,
		State:     state,
		CreatedAt: time.Now(),
	}
	codeStoreMutex.Unlock()

	logger.Info("Authorization code received", zap.String("code", code[:10]+"..."))

	// Редиректим на фронтенд с code и state
	// Токены НЕ передаются в URL - это безопасно!
	frontendURL := fmt.Sprintf(
		"%s/auth/callback?code=%s&state=%s",
		h.config.FrontendURL,
		code,
		state,
	)

	c.Redirect(http.StatusFound, frontendURL)
}

type exchangeTokenRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

type exchangeTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// @Summary Exchange Code for Tokens
// @Tags Auth
// @Description Обмен authorization code на access и refresh токены (как в Keycloak)
// @ModuleID exchangeToken
// @Accept  json
// @Produce  json
// @Param input body exchangeTokenRequest true "Code и State"
// @Success 200 {object} exchangeTokenResponse
// @Failure 400 {object} ErrorStruct
// @Router /users/auth/token [post]
func (h *Handler) exchangeToken(c *gin.Context) {
	var req exchangeTokenRequest
	if err := c.BindJSON(&req); err != nil {
		logger.Error("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	logger.Info("Token exchange request", zap.String("code", req.Code[:10]+"..."))

	// Проверяем, что code существует и валиден
	codeStoreMutex.RLock()
	codeData, exists := codeStore[req.Code]
	codeStoreMutex.RUnlock()

	if !exists {
		logger.Error("code not found")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code"})
		return
	}

	// Проверяем, что state совпадает
	if codeData.State != req.State {
		logger.Error("state mismatch",
			zap.String("expected", codeData.State),
			zap.String("received", req.State))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_state"})
		return
	}

	// Проверяем, что state еще валиден
	if !stateStore[req.State] {
		logger.Error("state expired or invalid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "state_expired"})
		return
	}

	// Проверяем срок действия code (10 минут)
	if time.Since(codeData.CreatedAt) > 10*time.Minute {
		logger.Error("code expired")
		codeStoreMutex.Lock()
		delete(codeStore, req.Code)
		codeStoreMutex.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_expired"})
		return
	}

	// Обменять code на токены через ESIA
	result, err := h.services.Users.Auth(
		c.Request.Context(),
		req.Code,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
	if err != nil {
		logger.Error("auth failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization_failed"})
		return
	}

	// Удаляем использованные code и state (одноразовое использование)
	codeStoreMutex.Lock()
	delete(codeStore, req.Code)
	codeStoreMutex.Unlock()

	delete(stateStore, req.State)

	logger.Info("Token exchange successful")

	// Возвращаем токены
	response := exchangeTokenResponse{
		AccessToken: result.AccessToken,
	}

	c.JSON(http.StatusOK, response)
}

// generateState генерирует случайный state для OAuth
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
