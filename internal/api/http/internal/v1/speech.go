package v1

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

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
// @Param audio formData file true "Аудиофайл в формате MP3"
// @Success 200 {object} speechResponse "Распознанный текст"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /speech/recognize [post]
// @Security UserAuth
func (h *Handler) parseAudioToText(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем файл из запроса
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		logger.Error("Failed to get audio file from request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось получить аудиофайл"})
		return
	}
	defer file.Close()

	// Читаем содержимое файла
	fileData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read audio file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось прочитать файл"})
		return
	}

	// Определяем MIME-тип на основе расширения файла
	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))

	// Gigachat принимает только MP3 формат
	if ext != ".mp3" {
		logger.Error("Unsupported audio format", zap.String("ext", ext))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Неподдерживаемый формат аудио: %s. Gigachat поддерживает только MP3", ext)})
		return
	}

	mimeType := "audio/mp3"

	logger.Info("Processing audio file",
		zap.String("filename", filename),
		zap.String("mime_type", mimeType),
		zap.Int("size", len(fileData)))

	// 1. Загружаем файл в хранилище GigaChat
	uploadResp, err := h.gigachatClient.UploadFile(ctx, fileData, filename, mimeType)
	if err != nil {
		logger.Error("Failed to upload file to GigaChat", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось загрузить файл в GigaChat: %v", err)})
		return
	}

	logger.Info("File uploaded successfully", zap.String("file_id", uploadResp.ID))

	// 2. Отправляем запрос на распознавание речи
	transcribedText, err := h.gigachatClient.TranscribeAudio(ctx, uploadResp.ID)
	if err != nil {
		logger.Error("Failed to transcribe audio", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось распознать речь: %v", err)})
		return
	}

	logger.Info("Audio transcribed successfully", zap.String("text", transcribedText))

	c.JSON(http.StatusOK, speechResponse{
		Text: transcribedText,
	})
}
