package v1

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
// @Param audio formData file true "Аудиофайл (MP3 или WebM)"
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

	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))

	var fileData []byte
	var mimeType string

	if ext == ".webm" {
		// Создаем временный файл для webm
		tmpWebm, err := os.CreateTemp("", "speech-*.webm")
		if err != nil {
			logger.Error("Failed to create temp file", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}
		defer os.Remove(tmpWebm.Name())
		defer tmpWebm.Close()

		// Сохраняем webm во временный файл
		if _, err := io.Copy(tmpWebm, file); err != nil {
			logger.Error("Failed to write webm to temp file", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}
		tmpWebm.Sync()

		// Путь для выходного mp3
		mp3Path := strings.TrimSuffix(tmpWebm.Name(), ".webm") + ".mp3"
		defer os.Remove(mp3Path)

		// Конвертируем с помощью ffmpeg
		cmd := exec.Command("ffmpeg", "-i", tmpWebm.Name(), "-vn", "-ar", "44100", "-ac", "2", "-b:a", "192k", "-y", mp3Path)
		if output, err := cmd.CombinedOutput(); err != nil {
			logger.Error("ffmpeg conversion failed", zap.Error(err), zap.String("output", string(output)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конвертации аудио"})
			return
		}

		// Читаем конвертированный mp3
		fileData, err = os.ReadFile(mp3Path)
		if err != nil {
			logger.Error("Failed to read converted mp3", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}

		filename = strings.TrimSuffix(filename, ".webm") + ".mp3"
		mimeType = "audio/mp3"

	} else if ext == ".mp3" {
		// Читаем содержимое файла
		fileData, err = io.ReadAll(file)
		if err != nil {
			logger.Error("Failed to read audio file", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось прочитать файл"})
			return
		}
		mimeType = "audio/mp3"
	} else {
		logger.Error("Unsupported audio format", zap.String("ext", ext))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Неподдерживаемый формат аудио: %s. Поддерживаются MP3 и WebM", ext)})
		return
	}

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
