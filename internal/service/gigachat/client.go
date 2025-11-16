package gigachat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/vibe-gaming/backend/internal/config"
	logger "github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type SpeechResponse struct {
	Result string `json:"result"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type EnhancedSearchResult struct {
	OriginalQuery string   `json:"original_query"`
	EnhancedTerms []string `json:"enhanced_terms"`
}

func NewClient(cfg config.GigachatConfig) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) RecognizeSpeech(ctx context.Context, audioData []byte, filename string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = part.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	if err = writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result SpeechResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil
}

// EnhanceSearchQuery улучшает поисковый запрос через GigaChat:
// - исправляет орфографические ошибки
// - добавляет морфологические варианты слов
// - находит синонимы
func (c *Client) EnhanceSearchQuery(ctx context.Context, query string) ([]string, error) {
	logger.Info("EnhanceSearchQuery called",
		zap.String("query", query),
		zap.String("baseURL", c.baseURL),
		zap.Bool("hasAPIKey", c.apiKey != ""))

	if query == "" {
		logger.Info("Empty query provided to EnhanceSearchQuery")
		return []string{query}, nil
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

	reqBody := ChatRequest{
		Model: "GigaChat",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("Failed to marshal GigaChat request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	requestURL := c.baseURL + "/v1/chat/completions"
	logger.Info("Sending request to GigaChat", zap.String("url", requestURL))

	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create GigaChat request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send request to GigaChat", zap.Error(err), zap.String("url", requestURL))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	logger.Info("Received response from GigaChat", zap.Int("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Error("GigaChat returned non-OK status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", string(bodyBytes)))
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		logger.Error("Failed to decode GigaChat response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
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
