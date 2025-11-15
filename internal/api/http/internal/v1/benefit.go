package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func (h *Handler) initBenefits(api *gin.RouterGroup) {
	benefits := api.Group("/benefits")
	{
		benefits.GET("", h.getBenefitsList)
		benefits.GET("/:id", h.getBenefitByID)
	}
}

type benefitResponse struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ValidFrom    string   `json:"valid_from"`
	ValidTo      string   `json:"valid_to"`
	Type         string   `json:"type"`
	TargetGroups []string `json:"target_groups"`
	Longitude    *float64 `json:"longitude,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Region       []int    `json:"region"`
	Requirement  string   `json:"requirement"`
	HowToUse     *string  `json:"how_to_use,omitempty"`
	SourceURL    string   `json:"source_url"`
}

type benefitsListResponse struct {
	Benefits []benefitResponse `json:"benefits"`
}

// @Summary Get Benefits List
// @Tags Benefits
// @Description Получить список всех льгот
// @ModuleID getBenefitsList
// @Accept  json
// @Produce  json
// @Success 200 {object} benefitsListResponse
// @Failure 500 {object} ErrorStruct
// @Router /benefits [get]
func (h *Handler) getBenefitsList(c *gin.Context) {
	benefits, err := h.services.Benefits.GetAll(c.Request.Context())
	if err != nil {
		logger.Error("failed to get benefits list", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get benefits list"})
		return
	}

	response := benefitsListResponse{
		Benefits: make([]benefitResponse, 0, len(benefits)),
	}

	for _, benefit := range benefits {
		targetGroups := make([]string, 0, len(benefit.TargetGroupIDs))
		for _, tg := range benefit.TargetGroupIDs {
			targetGroups = append(targetGroups, string(tg))
		}

		response.Benefits = append(response.Benefits, benefitResponse{
			ID:           benefit.ID.String(),
			Title:        benefit.Title,
			Description:  benefit.Description,
			ValidFrom:    benefit.ValidFrom.Format("2006-01-02"),
			ValidTo:      benefit.ValidTo.Format("2006-01-02"),
			Type:         benefit.Type,
			TargetGroups: targetGroups,
			Longitude:    benefit.Longitude,
			Latitude:     benefit.Latitude,
			Region:       benefit.Region,
			Requirement:  benefit.Requirement,
			HowToUse:     benefit.HowToUse,
			SourceURL:    benefit.SourceURL,
		})
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get Benefit By ID
// @Tags Benefits
// @Description Получить льготу по ID
// @ModuleID getBenefitByID
// @Accept  json
// @Produce  json
// @Param id path string true "Benefit ID"
// @Success 200 {object} benefitResponse
// @Failure 400 {object} ErrorStruct
// @Failure 404 {object} ErrorStruct
// @Failure 500 {object} ErrorStruct
// @Router /benefits/{id} [get]
func (h *Handler) getBenefitByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		logger.Error("benefit id is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "benefit id is required"})
		return
	}

	benefit, err := h.services.Benefits.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			logger.Error("benefit not found", zap.String("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "benefit not found"})
			return
		}
		logger.Error("failed to get benefit by id", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get benefit"})
		return
	}

	targetGroups := make([]string, 0, len(benefit.TargetGroupIDs))
	for _, tg := range benefit.TargetGroupIDs {
		targetGroups = append(targetGroups, string(tg))
	}

	response := benefitResponse{
		ID:           benefit.ID.String(),
		Title:        benefit.Title,
		Description:  benefit.Description,
		ValidFrom:    benefit.ValidFrom.Format("2006-01-02"),
		ValidTo:      benefit.ValidTo.Format("2006-01-02"),
		Type:         benefit.Type,
		TargetGroups: targetGroups,
		Longitude:    benefit.Longitude,
		Latitude:     benefit.Latitude,
		Region:       benefit.Region,
		Requirement:  benefit.Requirement,
		HowToUse:     benefit.HowToUse,
		SourceURL:    benefit.SourceURL,
	}

	c.JSON(http.StatusOK, response)
}
