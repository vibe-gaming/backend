package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vibe-gaming/backend/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	codes  map[string]*AuthCode
	tokens map[string]*Token
	mu     sync.RWMutex
}

type AuthCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	State       string
	CreatedAt   time.Time
}

type Token struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	TokenType    string
	CreatedAt    time.Time
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserInfo struct {
	OID           string   `json:"oid"`
	FirstName     string   `json:"firstName"`
	LastName      string   `json:"lastName"`
	MiddleName    string   `json:"middleName,omitempty"`
	BirthDate     string   `json:"birthDate"`
	Gender        string   `json:"gender"`
	SNILS         string   `json:"snils"`
	INN           string   `json:"inn,omitempty"`
	Email         string   `json:"email,omitempty"`
	Mobile        string   `json:"mobile,omitempty"`
	Trusted       bool     `json:"trusted"`
	Verified      bool     `json:"verified"`
	Citizenship   string   `json:"citizenship,omitempty"`
	Status        string   `json:"status"`
	Addresses     []string `json:"addresses,omitempty"`
	Documents     []string `json:"documents,omitempty"`
	Kids          []string `json:"kids,omitempty"`
	Organizations []string `json:"organizations,omitempty"`
}

func New() *Handler {
	return &Handler{
		codes:  make(map[string]*AuthCode),
		tokens: make(map[string]*Token),
	}
}

// OAuth2 Authorization endpoint
func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	logger.Info("Authorization request", zap.String("path", r.URL.Path))

	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")
	responseType := r.URL.Query().Get("response_type")

	logger.Debug("Authorization params",
		zap.String("client_id", clientID),
		zap.String("redirect_uri", redirectURI),
		zap.String("state", state),
		zap.String("scope", scope),
		zap.String("response_type", responseType),
	)

	if clientID == "" || redirectURI == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	code := h.generateCode()

	h.mu.Lock()
	h.codes[code] = &AuthCode{
		Code:        code,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		State:       state,
		CreatedAt:   time.Now(),
	}
	h.mu.Unlock()

	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += fmt.Sprintf("&state=%s", state)
	}

	logger.Info("Redirecting with code", zap.String("redirect_url", redirectURL))
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// OAuth2 Token endpoint
func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	logger.Info("Token request", zap.String("path", r.URL.Path))

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")

	logger.Debug("Token params",
		zap.String("grant_type", grantType),
		zap.String("code", code),
		zap.String("client_id", clientID),
		zap.String("redirect_uri", redirectURI),
	)

	if grantType != "authorization_code" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "unsupported_grant_type"})
		return
	}

	h.mu.RLock()
	authCode, exists := h.codes[code]
	h.mu.RUnlock()

	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		return
	}

	if authCode.ClientID != clientID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_client"})
		return
	}

	accessToken := h.generateToken()
	refreshToken := h.generateToken()
	idToken := h.generateIDToken()

	token := &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		CreatedAt:    time.Now(),
	}

	h.mu.Lock()
	h.tokens[accessToken] = token
	delete(h.codes, code)
	h.mu.Unlock()

	response := TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	logger.Info("Token issued", zap.String("access_token", accessToken[:10]+"..."))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// User info endpoint
func (h *Handler) UserInfo(w http.ResponseWriter, r *http.Request) {
	logger.Info("UserInfo request", zap.String("path", r.URL.Path))

	auth := r.Header.Get("Authorization")
	if auth == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	accessToken := parts[1]

	h.mu.RLock()
	_, exists := h.tokens[accessToken]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	// Возвращаем мок данные пользователя
	userInfo := UserInfo{
		OID:         "1000000001",
		FirstName:   "Иван",
		LastName:    "Иванов",
		MiddleName:  "Иванович",
		BirthDate:   "01.01.1990",
		Gender:      "M",
		SNILS:       "12345678901",
		INN:         "123456789012",
		Email:       "ivanov@example.com",
		Mobile:      "+79991234567",
		Trusted:     true,
		Verified:    true,
		Citizenship: "RUS",
		Status:      "REGISTERED",
	}

	logger.Info("UserInfo response", zap.String("oid", userInfo.OID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// Person by OID endpoint
func (h *Handler) GetPerson(w http.ResponseWriter, r *http.Request) {
	logger.Info("GetPerson request", zap.String("path", r.URL.Path))

	auth := r.Header.Get("Authorization")
	if auth == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	accessToken := parts[1]

	h.mu.RLock()
	_, exists := h.tokens[accessToken]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	// OID из пути URL, например /rs/prns/1000000001
	pathParts := strings.Split(r.URL.Path, "/")
	oid := pathParts[len(pathParts)-1]

	userInfo := UserInfo{
		OID:         oid,
		FirstName:   "Иван",
		LastName:    "Иванов",
		MiddleName:  "Иванович",
		BirthDate:   "01.01.1990",
		Gender:      "M",
		SNILS:       "12345678901",
		INN:         "123456789012",
		Email:       "ivanov@example.com",
		Mobile:      "+79991234567",
		Trusted:     true,
		Verified:    true,
		Citizenship: "RUS",
		Status:      "REGISTERED",
	}

	logger.Info("GetPerson response", zap.String("oid", userInfo.OID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

func (h *Handler) generateCode() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *Handler) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *Handler) generateIDToken() string {
	// Простой мок JWT токена
	header := base64.URLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.URLEncoding.EncodeToString([]byte(`{"sub":"1000000001","aud":"mock","iat":` + fmt.Sprintf("%d", time.Now().Unix()) + `,"exp":` + fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()) + `}`))
	signature := base64.URLEncoding.EncodeToString([]byte("mock_signature"))
	return fmt.Sprintf("%s.%s.%s", header, payload, signature)
}
