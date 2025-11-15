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
	codes        map[string]*AuthCode
	tokens       map[string]*Token
	userSessions map[string]string // code -> phone number
	mu           sync.RWMutex
}

type AuthCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	State       string
	PhoneNumber string
	CreatedAt   time.Time
}

type Token struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	TokenType    string
	PhoneNumber  string // Номер телефона пользователя
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
		codes:        make(map[string]*AuthCode),
		tokens:       make(map[string]*Token),
		userSessions: make(map[string]string),
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

	// Показываем форму для ввода номера телефона
	html := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Вход через Госуслуги</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #0d47a1 0%, #1976d2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
            max-width: 400px;
            width: 100%;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo-icon {
            width: 80px;
            height: 80px;
            background: #0d47a1;
            border-radius: 50%;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 40px;
            font-weight: bold;
            margin-bottom: 10px;
        }
        h1 {
            color: #333;
            font-size: 24px;
            text-align: center;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #666;
            text-align: center;
            font-size: 14px;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            color: #333;
            font-size: 14px;
            font-weight: 500;
            margin-bottom: 8px;
        }
        input[type="tel"] {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        input[type="tel"]:focus {
            outline: none;
            border-color: #1976d2;
        }
        button {
            width: 100%;
            padding: 14px;
            background: #0d47a1;
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.3s;
        }
        button:hover {
            background: #1976d2;
        }
        button:active {
            background: #0a3a7f;
        }
        .info {
            margin-top: 20px;
            padding: 12px;
            background: #e3f2fd;
            border-radius: 8px;
            font-size: 13px;
            color: #0d47a1;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <div class="logo-icon">Г</div>
        </div>
        <h1>Вход через Госуслуги</h1>
        <p class="subtitle">Введите номер телефона для авторизации</p>
        <form method="POST" action="/aas/oauth2/authorize">
            <input type="hidden" name="client_id" value="` + clientID + `">
            <input type="hidden" name="redirect_uri" value="` + redirectURI + `">
            <input type="hidden" name="state" value="` + state + `">
            <input type="hidden" name="scope" value="` + scope + `">
            <input type="hidden" name="response_type" value="` + responseType + `">
            
            <div class="form-group">
                <label for="phone">Номер телефона</label>
                <input 
                    type="tel" 
                    id="phone-display" 
                    placeholder="+7 (999) 123-45-67" 
                    required
                    autocomplete="tel"
                >
                <input type="hidden" id="phone" name="phone">
            </div>
            
            <button type="submit">Продолжить</button>
            
            <div class="info">
                🔒 Тестовая среда ЕСИА<br>
                Введите любой номер телефона
            </div>
        </form>
    </div>
    
    <script>
        const phoneInput = document.getElementById('phone-display');
        const phoneHidden = document.getElementById('phone');
        const form = document.querySelector('form');
        
        // Маска для отображения
        phoneInput.addEventListener('input', function(e) {
            let value = e.target.value.replace(/\D/g, '');
            
            // Ограничиваем до 11 цифр
            if (value.length > 11) {
                value = value.slice(0, 11);
            }
            
            let formattedValue = '';
            
            if (value.length > 0) {
                formattedValue = '+7';
                if (value.length > 1) {
                    formattedValue += ' (' + value.slice(1, 4);
                }
                if (value.length >= 4) {
                    formattedValue += ') ' + value.slice(4, 7);
                }
                if (value.length >= 7) {
                    formattedValue += '-' + value.slice(7, 9);
                }
                if (value.length >= 9) {
                    formattedValue += '-' + value.slice(9, 11);
                }
            }
            
            e.target.value = formattedValue;
        });
        
        // При отправке формы - форматируем в "79644223811"
        form.addEventListener('submit', function(e) {
            const displayValue = phoneInput.value.replace(/\D/g, '');
            
            // Проверяем, что введено 11 цифр
            if (displayValue.length !== 11) {
                e.preventDefault();
                alert('Пожалуйста, введите полный номер телефона');
                return false;
            }
            
            // Сохраняем в скрытое поле в формате "79644223811"
            phoneHidden.value = displayValue;
        });
        
        // Автофокус на поле
        phoneInput.focus();
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// AuthorizeSubmit обрабатывает отправку формы авторизации
func (h *Handler) AuthorizeSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	phoneNumber := r.FormValue("phone")

	logger.Info("Authorization form submitted",
		zap.String("client_id", clientID),
		zap.String("phone", phoneNumber),
	)

	if clientID == "" || redirectURI == "" || phoneNumber == "" {
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
		PhoneNumber: phoneNumber,
		CreatedAt:   time.Now(),
	}
	h.userSessions[code] = phoneNumber
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
		PhoneNumber:  authCode.PhoneNumber,
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
	token, exists := h.tokens[accessToken]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	// Используем номер телефона из токена
	phoneNumber := token.PhoneNumber
	if phoneNumber == "" {
		phoneNumber = "+79991234567" // дефолтный номер если не задан
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
		Mobile:      phoneNumber,
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
