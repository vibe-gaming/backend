package gigachat

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

const (
	baseURLToken = "https://ngw.devices.sberbank.ru:9443/api/v2"
	baseURL      = "https://gigachat.devices.sberbank.ru/api/v1"
)

// Client представляет клиент для работы с API GigaChat
type Client struct {
	clientID    string
	basicAuth   string
	scope       string
	httpClient  *http.Client
	token       *TokenResponse
	tokenExpiry time.Time
}

// NewClient создает новый клиент GigaChat
func NewClient(authoizationKey string) *Client {
	// Создаем базовую авторизацию
	return &Client{
		basicAuth: fmt.Sprintf("Basic %s", authoizationKey),
		scope:     ScopePersonal, // По умолчанию используем персональный scope
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// SetScope устанавливает scope для клиента
func (c *Client) SetScope(scope string) {
	c.scope = scope
}

// SetClientID устанавливает client ID для клиента
func (c *Client) SetClientID(clientID string) {
	c.clientID = clientID
}

// GetToken получает токен доступа
func (c *Client) GetToken() (*TokenResponse, error) {
	dataScope := fmt.Sprintf("scope=%s", c.scope)

	// Создаем запрос
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth", baseURLToken), bytes.NewBufferString(dataScope))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", c.basicAuth)
	httpReq.Header.Set("RqUID", generateUUID())

	// Отправляем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("ошибка разбора ответа об ошибке: %w", err)
		}
		return nil, &APIError{
			Code:    errResp.Code,
			Message: errResp.Message,
		}
	}

	// Разбираем успешный ответ
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	// Сохраняем токен и время его истечения
	c.token = &tokenResp
	c.tokenExpiry = time.Unix(tokenResp.ExpiresAt/1000-60, 0)

	return &tokenResp, nil
}

// ensureValidToken проверяет и обновляет токен при необходимости
func (c *Client) ensureValidToken() error {
	if c.token == nil || time.Now().After(c.tokenExpiry) {
		_, err := c.GetToken()
		if err != nil {
			return fmt.Errorf("ошибка обновления токена: %w", err)
		}
	}
	return nil
}

// SendBytes отправляет байты в чат
func (c *Client) SendBytes(query []byte) ([]byte, error) {
	// Проверяем валидность токена
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	// Создаем запрос
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", baseURL), bytes.NewBuffer(query))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))
	httpReq.Header.Set("X-Client-ID", c.clientID)
	httpReq.Header.Set("X-Request-ID", generateUUID())
	httpReq.Header.Set("X-Session-ID", generateUUID())

	// Отправляем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
		} else {
			return body, fmt.Errorf("некорректный статус ответа: %v", resp.Status)
		}
	}

	// Возвращаем тело ответа
	return io.ReadAll(resp.Body)
}

// Chat отправляет запрос к чату
func (c *Client) Chat(req *ChatRequest) (*ChatResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	body, err := c.SendBytes(jsonData)
	if err != nil {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("ошибка разбора ответа об ошибке: %w", err)
		}
		return nil, &APIError{
			Code:    errResp.Code,
			Message: errResp.Message,
		}
	}

	// Разбираем успешный ответ
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return &chatResp, nil
}

// generateUUID генерирует UUID для RqUID
func generateUUID() string {
	// Генерация реального UUID
	return uuid.New().String()
}

type EnhancedSearchResult struct {
	OriginalQuery string   `json:"original_query"`
	EnhancedTerms []string `json:"enhanced_terms"`
}

// EnhanceSearchQuery улучшает поисковый запрос через GigaChat:
// - исправляет орфографические ошибки
// - добавляет морфологические варианты слов
// - находит синонимы
func (c *Client) EnhanceSearchQuery(ctx context.Context, query string) ([]string, error) {
	logger.Info("EnhanceSearchQuery called",
		zap.String("query", query),
		zap.String("baseURL", baseURL),
		zap.Bool("hasAPIKey", c.basicAuth != ""))

	if query == "" {
		logger.Info("Empty query provided to EnhanceSearchQuery")
		return []string{query}, nil
	}

	// Проверяем и обновляем токен при необходимости
	if err := c.ensureValidToken(); err != nil {
		logger.Error("Failed to ensure valid token", zap.Error(err))
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Формируем промпт для GigaChat
	prompt := fmt.Sprintf(`Проанализируй поисковый запрос и помоги улучшить его для поиска социальных льгот и мер поддержки.

Запрос: "%s"

Твоя задача:
1. Исправь орфографические ошибки (если есть)
2. Добавь морфологические варианты слов (например, "студент" -> "студенты", "студентам", "студентов")
3. Найди синонимы и связанные термины (например, "школьник" -> "ученик", "учащийся")

Верни результат в виде списка поисковых терминов через запятую, БЕЗ объяснений и дополнительного текста.
Каждый термин должен быть коротким (1-3 слова максимум).

Пример формата ответа:
студент, студенты, студентам, учащийся, обучающийся

Не добавляй нумерацию, точки или другое форматирование, только термины через запятую.`, query)

	reqBody := &ChatRequest{
		Model: "GigaChat",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	logger.Info("Sending chat request to GigaChat")
	chatResp, err := c.Chat(reqBody)
	if err != nil {
		logger.Error("Failed to send chat request", zap.Error(err))
		return nil, fmt.Errorf("failed to send chat request: %w", err)
	}

	logger.Info("Successfully decoded GigaChat response", zap.Int("choices_count", len(chatResp.Choices)))

	if len(chatResp.Choices) == 0 {
		logger.Info("GigaChat returned no choices, using original query")
		return []string{query}, nil
	}

	// Парсим ответ от GigaChat
	responseText := chatResp.Choices[0].Message.Content
	logger.Info("GigaChat response text", zap.String("response", responseText))

	// Разбиваем по запятой и очищаем от лишних пробелов
	terms := []string{}
	rawTerms := strings.Split(responseText, ",")

	for _, term := range rawTerms {
		cleaned := strings.TrimSpace(term)
		// Убираем возможные артефакты (нумерацию, точки и т.д.)
		cleaned = strings.TrimPrefix(cleaned, "-")
		cleaned = strings.TrimPrefix(cleaned, "•")
		cleaned = strings.TrimSpace(cleaned)

		if cleaned != "" && len(cleaned) > 1 {
			terms = append(terms, cleaned)
		}
	}

	logger.Info("Parsed terms from GigaChat", zap.Strings("terms", terms), zap.Int("count", len(terms)))

	// Если не удалось распарсить ответ, возвращаем оригинальный запрос
	if len(terms) == 0 {
		logger.Info("Failed to parse any terms from GigaChat response, using original query")
		return []string{query}, nil
	}

	logger.Info("Successfully enhanced search query",
		zap.String("original", query),
		zap.Strings("enhanced", terms))
	return terms, nil
}

// UploadFile загружает аудиофайл в хранилище GigaChat
func (c *Client) UploadFile(ctx context.Context, fileData []byte, filename, mimeType string) (*FileUploadResponse, error) {
	// Логируем сигнатуру файла для отладки (первые 16 байт)
	signature := ""
	if len(fileData) >= 16 {
		signature = fmt.Sprintf("%X", fileData[:16])
	}

	logger.Info("UploadFile called",
		zap.String("filename", filename),
		zap.String("mime_type", mimeType),
		zap.Int("size", len(fileData)),
		zap.String("signature", signature))

	// Проверяем и обновляем токен при необходимости
	if err := c.ensureValidToken(); err != nil {
		logger.Error("Failed to ensure valid token", zap.Error(err))
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Создаем multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Очищаем MIME-тип от параметров (например, ;codecs=opus)
	cleanMimeType := mimeType
	if idx := strings.Index(mimeType, ";"); idx > 0 {
		cleanMimeType = mimeType[:idx]
	}

	logger.Info("Using cleaned MIME type", zap.String("original", mimeType), zap.String("clean", cleanMimeType))

	// Создаем часть с правильным MIME-типом
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename)}
	if cleanMimeType != "" {
		h["Content-Type"] = []string{cleanMimeType}
	}

	part, err := writer.CreatePart(h)
	if err != nil {
		logger.Error("Failed to create form part", zap.Error(err))
		return nil, fmt.Errorf("failed to create form part: %w", err)
	}

	if _, err := part.Write(fileData); err != nil {
		logger.Error("Failed to write file data", zap.Error(err))
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	// Добавляем purpose (опционально)
	if err := writer.WriteField("purpose", "general"); err != nil {
		logger.Error("Failed to write purpose field", zap.Error(err))
		return nil, fmt.Errorf("failed to write purpose field: %w", err)
	}

	if err := writer.Close(); err != nil {
		logger.Error("Failed to close writer", zap.Error(err))
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Создаем запрос
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/files", baseURL), body)
	if err != nil {
		logger.Error("Failed to create request", zap.Error(err))
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))
	httpReq.Header.Set("X-Client-ID", c.clientID)
	httpReq.Header.Set("X-Request-ID", generateUUID())

	// Отправляем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("Failed to send request", zap.Error(err))
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response", zap.Error(err))
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус
	if resp.StatusCode != http.StatusOK {
		logger.Error("Upload failed", zap.Int("status", resp.StatusCode), zap.String("body", string(responseBody)))
		return nil, fmt.Errorf("upload failed: %s: %s", resp.Status, string(responseBody))
	}

	// Разбираем ответ
	var uploadResp FileUploadResponse
	if err := json.Unmarshal(responseBody, &uploadResp); err != nil {
		logger.Error("Failed to unmarshal response", zap.Error(err), zap.String("body", string(responseBody)))
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	logger.Info("Successfully uploaded file", zap.String("file_id", uploadResp.ID))
	return &uploadResp, nil
}

// TranscribeAudio распознает речь из аудиофайла используя chat/completions с attachments
func (c *Client) TranscribeAudio(ctx context.Context, fileID string) (string, error) {
	logger.Info("TranscribeAudio called", zap.String("file_id", fileID))

	// Проверяем и обновляем токен при необходимости
	if err := c.ensureValidToken(); err != nil {
		logger.Error("Failed to ensure valid token", zap.Error(err))
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Формируем запрос к чату с прикрепленным файлом
	reqBody := &ChatRequest{
		Model: ModelGigaChatPro,
		Messages: []Message{
			{
				Role:        RoleUser,
				Content:     "Распознай речь из прикреплённого аудиофайла и верни только текст, который был произнесён, без лишних пояснений.",
				Attachments: []string{fileID},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("Failed to marshal request", zap.Error(err))
		return "", fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	logger.Info("Sending transcription request", zap.String("request_json", string(jsonData)))

	// Создаем запрос
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", baseURL), bytes.NewReader(jsonData))
	if err != nil {
		logger.Error("Failed to create request", zap.Error(err))
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))
	httpReq.Header.Set("X-Client-ID", c.clientID)
	httpReq.Header.Set("X-Request-ID", generateUUID())
	httpReq.Header.Set("X-Session-ID", generateUUID())

	// Отправляем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("Failed to send request", zap.Error(err))
		return "", fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response", zap.Error(err))
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	logger.Info("Received transcription response", zap.String("response_json", string(responseBody)))

	// Проверяем статус
	if resp.StatusCode != http.StatusOK {
		logger.Error("Transcription request failed", zap.Int("status", resp.StatusCode), zap.String("body", string(responseBody)))
		return "", fmt.Errorf("transcription request failed: %s: %s", resp.Status, string(responseBody))
	}

	// Разбираем ответ
	var chatResp ChatResponse
	if err := json.Unmarshal(responseBody, &chatResp); err != nil {
		logger.Error("Failed to unmarshal response", zap.Error(err), zap.String("body", string(responseBody)))
		return "", fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	// Извлекаем текст из ответа
	if len(chatResp.Choices) == 0 {
		logger.Error("No choices in response")
		return "", fmt.Errorf("нет вариантов ответа")
	}

	transcribedText := chatResp.Choices[0].Message.Content
	logger.Info("Successfully transcribed audio", zap.String("text", transcribedText))

	return transcribedText, nil
}
