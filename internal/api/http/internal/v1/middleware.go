package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userId"
)

func (h *Handler) userIdentityMiddleware(c *gin.Context) {
	id, err := h.parseAuthHeader(c)
	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			logger.Error("parse auth header failed", zap.Error(err))
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	c.Set(userCtx, id)
}

// optionalUserIdentityMiddleware пытается авторизовать пользователя, но не требует обязательной авторизации
// Используется для эндпоинтов, которые работают и с авторизацией, и без неё
func (h *Handler) optionalUserIdentityMiddleware(c *gin.Context) {
	authHeader := c.GetHeader(authorizationHeader)
	logger.Info("optionalUserIdentityMiddleware called",
		zap.String("path", c.Request.URL.Path),
		zap.Bool("has_auth_header", authHeader != ""))

	id, err := h.parseAuthHeader(c)
	if err == nil {
		// Если токен валидный - устанавливаем userId в контекст
		c.Set(userCtx, id)
	}
	// Если ошибка - просто продолжаем без установки userId
	c.Next()
}

func (h *Handler) parseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	slog.String("header", header)
	if header == "" {
		return "", errors.New("empty auth header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}

	if len(headerParts[1]) == 0 {
		return "", errors.New("token is empty")
	}

	return h.tokenManager.Parse(headerParts[1])
}

func (h *Handler) getUserUUID(c *gin.Context) (uuid.UUID, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		return uuid.Nil, errors.New("user id not found")
	}

	logger.Info("user id found", zap.String("user_id", id.(string)))

	return uuid.MustParse(id.(string)), nil
}
