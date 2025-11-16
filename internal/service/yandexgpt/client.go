package yandexgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const sttEndpoint = "https://stt.api.cloud.yandex.net/speech/v1/stt:recognize"

// Client предоставляет методы для работы с распознаванием речи Яндекс GPT (SpeechKit STT)
type Client struct {
	apiKey     string
	folderID   string
	lang       string
	httpClient *http.Client
}

type sttResponse struct {
	Result       string `json:"result"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// NewClient создает новый клиент Yandex GPT STT
func NewClient(apiKey, folderID, lang string) *Client {
	if lang == "" {
		lang = "ru-RU"
	}

	return &Client{
		apiKey:   apiKey,
		folderID: folderID,
		lang:     lang,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Recognize выполняет синхронное распознавание речи
func (c *Client) Recognize(ctx context.Context, audio []byte, mimeType string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("yandex GPT api key is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sttEndpoint, bytes.NewReader(audio))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	query := req.URL.Query()
	if c.folderID != "" {
		query.Set("folderId", c.folderID)
	}
	query.Set("lang", c.lang)
	query.Set("topic", "general")
	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	if mimeType == "" {
		mimeType = "audio/wav"
	}
	req.Header.Set("Content-Type", mimeType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("yandex STT request failed: %s: %s", resp.Status, string(body))
	}

	var sttResp sttResponse
	if err := json.Unmarshal(body, &sttResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if sttResp.ErrorCode != "" {
		return "", fmt.Errorf("yandex STT error %s: %s", sttResp.ErrorCode, sttResp.ErrorMessage)
	}

	if sttResp.Result == "" {
		return "", fmt.Errorf("empty result received from yandex STT")
	}

	return sttResp.Result, nil
}
