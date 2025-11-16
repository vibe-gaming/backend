package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func (h *Handler) initCitiesRoutes(api *gin.RouterGroup) {
	cities := api.Group("/cities")
	cities.GET("", h.getCities)
}

// @Summary Get Cities
// @Tags Cities
// @Description Get all cities
// @ModuleID getCities
// @Accept  json
// @Produce  json
// @Success 200 {object} []domain.City
// @Failure 400 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /cities [get]
func (h *Handler) getCities(c *gin.Context) {
	cities, err := h.services.Cities.GetAll(c.Request.Context())
	if err != nil {
		logger.Error("get cities failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, cities)
}
