package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func (h *Handler) initOrganizationsRoutes(api *gin.RouterGroup) {
	organizations := api.Group("/organizations")
	{
		organizations.POST("", h.createOrganization)
		organizations.GET("", h.getOrganizations)
		organizations.GET("/:id", h.getOrganizationByID)
		organizations.PUT("/:id", h.updateOrganization)
		organizations.DELETE("/:id", h.deleteOrganization)
	}
}

type createOrganizationRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type createOrganizationResponse struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

// @Summary Get Organizations
// @Tags Organizations
// @Description Get all organizations, optionally filtered by city_id
// @ModuleID getOrganizations
// @Accept  json
// @Produce  json
// @Param city_id query string false "City ID (UUID) to filter organizations"
// @Success 200 {object} []domain.Organization
// @Failure 400 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /organizations [get]
func (h *Handler) getOrganizations(c *gin.Context) {
	cityID := c.Query("city_id")
	
	var organizations []domain.Organization
	var err error
	
	if cityID != "" {
		organizations, err = h.services.Organizations.GetAllByCityID(c.Request.Context(), cityID)
	} else {
		organizations, err = h.services.Organizations.GetAll(c.Request.Context())
	}
	
	if err != nil {
		logger.Error("get organizations failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, organizations)
}

// @Summary Create Organization
// @Tags Organizations
// @Description Создать новую организацию
// @ModuleID createOrganization
// @Accept  json
// @Produce  json
// @Param input body createOrganizationRequest true "Данные организации"
// @Success 201 {object} createOrganizationResponse
// @Failure 400 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /organizations [post]
func (h *Handler) createOrganization(c *gin.Context) {
	var req createOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	organization := &domain.Organization{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.services.Organizations.Create(c.Request.Context(), organization); err != nil {
		logger.Error("failed to create organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	response := createOrganizationResponse{
		ID:        organization.ID.String(),
		CreatedAt: organization.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusCreated, response)
}

// @Summary Get Organization By ID
// @Tags Organizations
// @Description Получить организацию по ID
// @ModuleID getOrganizationByID
// @Accept  json
// @Produce  json
// @Param id path string true "Organization ID (UUID)"
// @Success 200 {object} domain.Organization
// @Failure 400 {object} ErrorStruct
// @Failure 404 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /organizations/{id} [get]
func (h *Handler) getOrganizationByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("organization id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization id is required"})
		return
	}

	organization, err := h.services.Organizations.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Error("organization not found", zap.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		logger.Error("failed to get organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get organization"})
		return
	}

	c.JSON(http.StatusOK, organization)
}

// @Summary Update Organization
// @Tags Organizations
// @Description Обновить существующую организацию
// @ModuleID updateOrganization
// @Accept  json
// @Produce  json
// @Param id path string true "Organization ID (UUID)"
// @Param input body createOrganizationRequest true "Данные организации"
// @Success 200 {object} createOrganizationResponse
// @Failure 400 {object} ErrorStruct
// @Failure 404 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /organizations/{id} [put]
func (h *Handler) updateOrganization(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("organization id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization id is required"})
		return
	}

	// Получаем существующую организацию
	existingOrganization, err := h.services.Organizations.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Error("organization not found", zap.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		logger.Error("failed to get organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get organization"})
		return
	}

	var req createOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Обновление объекта Organization
	existingOrganization.Name = req.Name
	existingOrganization.Description = req.Description

	// Обновление организации через сервис
	if err := h.services.Organizations.Update(c.Request.Context(), existingOrganization); err != nil {
		logger.Error("failed to update organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	response := createOrganizationResponse{
		ID:        existingOrganization.ID.String(),
		CreatedAt: existingOrganization.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Delete Organization
// @Tags Organizations
// @Description Удалить организацию (soft delete)
// @ModuleID deleteOrganization
// @Accept  json
// @Produce  json
// @Param id path string true "Organization ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorStruct
// @Failure 404 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /organizations/{id} [delete]
func (h *Handler) deleteOrganization(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("organization id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization id is required"})
		return
	}

	err := h.services.Organizations.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Error("organization not found", zap.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		logger.Error("failed to delete organization", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}

