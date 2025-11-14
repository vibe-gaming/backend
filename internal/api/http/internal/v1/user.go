package v1

import (
	"errors"
	"net/http"

	"github.com/vibe-gaming/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")

	users.POST("/register", h.userRegister)
	users.POST("/auth", h.userAuth)
	users.POST("/auth/refresh", h.userRefresh)

	users.GET("/pong", h.userIdentityMiddleware, h.pong)
}

type userRegisterRequest struct {
	Login    string `json:"name" binding:"required,min=2,max=64" example:"wazzup"`
	Email    string `json:"email" binding:"required,email,max=64" example:"mail@mail.com"`
	Password string `json:"password" binding:"required,min=8,max=64" example:"notasecretpassword"`
}

// @Summary Регистрация
// @Tags User Auth
// @Description Создание аккаунта юзера
// @ModuleID userRegister
// @Accept  json
// @Produce  json
// @Param input body userRegisterRequest true "Регистрация"
// @Success 201
// @Failure 400 {object} ErrorStruct
// @Failure 500
// @Router /users/register [post]
func (h *Handler) userRegister(c *gin.Context) {
	var req userRegisterRequest
	if err := c.BindJSON(&req); err != nil {
		validationErrorResponse(c, err)
		return
	}

	err := h.services.Users.Register(c.Request.Context(), &service.UserRegisterInput{
		Login:    req.Login,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExist) {
			errorResponse(c, UserAlreadyExistsCode)
			return
		}
		h.logger.Error("failed to create user", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusCreated)
}

type userAuthRequest struct {
	Email    string `json:"email" binding:"required,email,max=64" example:"mail@mail.com"`
	Password string `json:"password" binding:"required,min=8,max=64" example:"notasecretpassword"`
}

type userAuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken uuid.UUID `json:"refresh_token"`
}

// @Summary Аутентификация
// @Tags User Auth
// @Description Аутентификация пользователей
// @ModuleID user
// @Accept  json
// @Produce  json
// @Param input body userAuthRequest true "Аутентификация"
// @Success 200 {object} userAuthResponse
// @Failure 400 {object} ErrorStruct
// @Router /users/auth [post]
func (h *Handler) userAuth(c *gin.Context) {
	var req userAuthRequest
	if err := c.BindJSON(&req); err != nil {
		validationErrorResponse(c, err)
	}

	result, err := h.services.Users.Auth(c.Request.Context(), &service.UserAuthInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: c.Request.UserAgent(),
		IP:        c.ClientIP(),
	})
	if err != nil {
		h.logger.Error("user auth failed", "error", err)
		c.Status(http.StatusBadRequest)
		return
	}

	response := userAuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) userRefresh(c *gin.Context) {

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
