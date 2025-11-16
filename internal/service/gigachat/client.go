package gigachat

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
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
			return body, fmt.Errorf("Некорректный статус ответа: %v", resp.Status)
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
