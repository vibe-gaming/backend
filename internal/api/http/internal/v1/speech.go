package v1

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

func (h *Handler) initSpeechRoutes(api *gin.RouterGroup) {
	speech := api.Group("/speech")

	speech.POST("/recognize", h.userIdentityMiddleware, h.parseAudioToText)
}

type speechResponse struct {
	Text string `json:"text"`
}

// @Summary Распознавание речи
// @Tags Speech
// @Description Распознает речь из аудиофайла и возвращает текст
// @Security UserAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Аудиофайл (поддерживаются форматы: wav, mp3, ogg)"
// @Success 200 {object} speechResponse "Распознанный текст"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /speech/recognize [post]
func (h *Handler) parseAudioToText(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		logger.Error("file not found", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не найден"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		logger.Error("failed to open file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer openedFile.Close()

	audioData, err := io.ReadAll(openedFile)
	if err != nil {
		logger.Error("failed to read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось прочитать файл"})
		return
	}

	text, err := h.gigachatClient.RecognizeSpeech(c.Request.Context(), audioData, file.Filename)
	if err != nil {
		logger.Error("failed to recognize speech", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось распознать речь"})
		return
	}

	c.JSON(http.StatusOK, speechResponse{
		Text: text,
	})
}
