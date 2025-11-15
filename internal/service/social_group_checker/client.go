package socialgroupchecker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SocialGroup представляет тип социальной группы
type SocialGroup string

const (
	Pensioners    SocialGroup = "pensioners"     // Пенсионеры
	Disabled      SocialGroup = "disabled"       // Инвалиды
	YoungFamilies SocialGroup = "young_families" // Молодые семьи
	LowIncome     SocialGroup = "low_income"     // Малоимущие
	Students      SocialGroup = "students"       // Студенты
	LargeFamilies SocialGroup = "large_families" // Многодетные семьи
	Children      SocialGroup = "children"       // Дети
	Veterans      SocialGroup = "veterans"       // Ветераны
)

// GroupStatus представляет статус проверки социальной группы
type GroupStatus string

const (
	StatusConfirmed GroupStatus = "подтвержден"
	StatusRejected  GroupStatus = "отклонен"
)

// CheckRequest представляет запрос на проверку социальных групп
type CheckRequest struct {
	SNILS  string        `json:"snils"`
	Groups []SocialGroup `json:"groups"`
}

// GroupResult представляет результат проверки одной социальной группы
type GroupResult struct {
	Group  SocialGroup `json:"group"`
	Status GroupStatus `json:"status"`
}

// CheckResponse представляет ответ на проверку социальных групп
type CheckResponse struct {
	SNILS   string        `json:"snils"`
	Results []GroupResult `json:"results"`
}

// HealthResponse представляет ответ health-check эндпоинта
type HealthResponse struct {
	Status string `json:"status"`
}

// Client представляет клиент для мок-сервиса проверки социальных групп
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создает новый клиент для мок-сервиса
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckGroups проверяет статус социальных групп для указанного СНИЛС
func (c *Client) CheckGroups(ctx context.Context, snils string, groups []SocialGroup) (*CheckResponse, error) {
	if snils == "" {
		return nil, fmt.Errorf("СНИЛС не может быть пустым")
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("список групп не может быть пустым")
	}

	reqBody := CheckRequest{
		SNILS:  snils,
		Groups: groups,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/check", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("неожиданный статус код: %d, тело: %s", resp.StatusCode, string(body))
	}

	var checkResp CheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		return nil, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	return &checkResp, nil
}

// IsValidGroup проверяет, является ли группа валидной
func IsValidGroup(group SocialGroup) bool {
	validGroups := map[SocialGroup]bool{
		Pensioners:    true,
		Disabled:      true,
		YoungFamilies: true,
		LowIncome:     true,
		Students:      true,
		LargeFamilies: true,
		Children:      true,
		Veterans:      true,
	}
	return validGroups[group]
}
