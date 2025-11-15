package esia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

type Client struct {
	config     config.ESIAConfig
	httpClient *http.Client
}

func NewClient(cfg config.ESIAConfig) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAuthorizationURL возвращает URL для перенаправления пользователя на ESIA для авторизации
func (c *Client) GetAuthorizationURL(state string) string {
	params := url.Values{}
	params.Add("client_id", c.config.ClientID)
	params.Add("redirect_uri", c.config.RedirectURI)
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("scope", c.config.Scope)

	return fmt.Sprintf("%s/aas/oauth2/ac?%s", c.config.BaseURL, params.Encode())
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// ExchangeCodeForToken обменивает authorization code на access token
func (c *Client) ExchangeCodeForToken(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", c.config.ClientID)
	data.Set("redirect_uri", c.config.RedirectURI)

	req, err := http.NewRequest("POST", c.config.BaseURL+"/aas/oauth2/te", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	logger.Debug("Exchanging code for token", zap.String("url", req.URL.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	logger.Debug("Token received", zap.String("access_token", tokenResp.AccessToken[:10]+"..."))

	return &tokenResp, nil
}

type UserInfo struct {
	OID         string `json:"oid"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	MiddleName  string `json:"middleName,omitempty"`
	BirthDate   string `json:"birthDate"`
	Gender      string `json:"gender"`
	SNILS       string `json:"snils"`
	INN         string `json:"inn,omitempty"`
	Email       string `json:"email,omitempty"`
	Mobile      string `json:"mobile,omitempty"`
	Trusted     bool   `json:"trusted"`
	Verified    bool   `json:"verified"`
	Citizenship string `json:"citizenship,omitempty"`
	Status      string `json:"status"`
}

// GetUserInfo получает информацию о пользователе используя access token
func (c *Client) GetUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", c.config.BaseURL+"/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	logger.Debug("Getting user info", zap.String("url", req.URL.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	logger.Debug("User info received", zap.String("oid", userInfo.OID))

	return &userInfo, nil
}
