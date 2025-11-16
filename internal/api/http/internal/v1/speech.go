package v1

import (
	"io"
	"net/http"
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
// @Param audio formData file true "Аудиофайл (поддерживаются форматы: ogg (opus), mp3, wav)"
// @Success 200 {object} speechResponse "Распознанный текст"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /speech/recognize [post]
func (h *Handler) parseAudioToText(c *gin.Context) {
	// Получаем файл из запроса
	file, err := c.FormFile("audio")
	if err != nil {
		logger.Error("file not found", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не найден"})
		return
	}

	// Получаем заголовок Content-Type из multipart формы
	fileHeader := file.Header.Get("Content-Type")

	logger.Info("Received audio file for transcription",
		zap.String("filename", file.Filename),
		zap.String("content_type", fileHeader),
		zap.Int64("size", file.Size))

	// Открываем файл
	openedFile, err := file.Open()
	if err != nil {
		logger.Error("failed to open file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer openedFile.Close()

	// Читаем данные файла
	audioData, err := io.ReadAll(openedFile)
	if err != nil {
		logger.Error("failed to read file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось прочитать файл"})
		return
	}

	// Определяем MIME-тип
	mimeType := detectAudioMimeType(audioData, fileHeader)
	logger.Info("Detected MIME type", zap.String("mime_type", mimeType))

	canonicalMime, ok := normalizeAudioMime(mimeType)
	if !ok {
		logger.Info("Unsupported audio format",
			zap.String("detected_mime", mimeType),
			zap.String("filename", file.Filename))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Поддерживаются только аудиофайлы в формате OGG (Opus), MP3 или WAV",
		})
		return
	}

	// Распознаем речь через Yandex GPT (SpeechKit STT)
	logger.Info("Starting audio transcription via Yandex GPT")
	transcribedText, err := h.yandexClient.Recognize(c.Request.Context(), audioData, canonicalMime)
	if err != nil {
		logger.Error("failed to transcribe audio via Yandex GPT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Яндекс GPT не смог распознать речь"})
		return
	}

	logger.Info("Successfully transcribed audio",
		zap.String("text", transcribedText),
		zap.Int("text_length", len(transcribedText)))

	// Возвращаем результат
	c.JSON(http.StatusOK, speechResponse{
		Text: transcribedText,
	})
}

// detectAudioMimeType определяет MIME-тип аудиофайла по сигнатуре и заголовку
func detectAudioMimeType(data []byte, headerContentType string) string {
	// Проверяем сигнатуры файлов
	if len(data) >= 4 {
		// WAV файл: RIFF....WAVE
		if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
			if len(data) >= 12 && data[8] == 0x57 && data[9] == 0x41 && data[10] == 0x56 && data[11] == 0x45 {
				return "audio/wav"
			}
		}

		// MP3 файл: ID3 или 0xFF 0xFB
		if (data[0] == 0x49 && data[1] == 0x44 && data[2] == 0x33) ||
			(data[0] == 0xFF && (data[1] == 0xFB || data[1] == 0xF3 || data[1] == 0xF2)) {
			return "audio/mpeg"
		}

		// OGG файл: OggS
		if data[0] == 0x4F && data[1] == 0x67 && data[2] == 0x67 && data[3] == 0x53 {
			return "audio/ogg"
		}
	}

	if len(data) >= 12 {
		// WebM файл
		if data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
			return "audio/webm"
		}
	}

	// Если есть заголовок Content-Type, используем его
	if headerContentType != "" && headerContentType != "application/octet-stream" {
		return headerContentType
	}

	// По умолчанию возвращаем WAV
	return "audio/wav"
}
func normalizeAudioMime(mime string) (string, bool) {
	m := strings.ToLower(strings.TrimSpace(mime))
	switch m {
	case "audio/wav", "audio/x-wav", "audio/wave", "audio/x-pn-wav":
		return "audio/x-wav", true
	case "audio/mpeg", "audio/mp3":
		return "audio/mpeg", true
	case "audio/ogg", "audio/x-ogg", "application/ogg":
		return "audio/ogg", true
	default:
		return "", false
	}
}
