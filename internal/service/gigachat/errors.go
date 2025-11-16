package gigachat

import "fmt"

// APIError представляет ошибку API
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %d - %s", e.Code, e.Message)
}

// ErrInvalidCredentials ошибка неверных учетных данных
var ErrInvalidCredentials = &APIError{
	Code:    401,
	Message: "Неверные учетные данные",
}

// ErrInvalidRequest ошибка некорректного запроса
var ErrInvalidRequest = &APIError{
	Code:    400,
	Message: "Некорректный формат запроса",
}

// ErrServerError ошибка сервера
var ErrServerError = &APIError{
	Code:    500,
	Message: "Внутренняя ошибка сервера",
}
