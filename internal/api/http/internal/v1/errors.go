package v1

// Errors
const (
	UnknownErrorCode    = 0
	UnknownErrorMessage = "unknown error"

	UserAlreadyExistsCode                 = 1001
	UserAlreadyExistsMessage              = "user already exists"
	UserNotFoundCode                      = 1002
	UserNotFoundMessage                   = "user not found"
	UserRefreshTokenCookieNotFoundCode    = 1003
	UserRefreshTokenCookieNotFoundMessage = "user refresh token cookie not found"
	UserRefreshTokenExpiredCode           = 1004
	UserRefreshTokenExpiredMessage        = "user refresh token expired"
)

type ErrorCode int
type ErrorMessage string

type ErrorStruct struct {
	ErrorCode    `json:"error_code"`
	ErrorMessage `json:"error_message"`
} // @name ErrorStruct

type ValidationErrorStruct struct {
	ErrorCode    int               `json:"error_code"`
	ErrorMessage string            `json:"error_message"`
	Errors       []ValidationError `json:"validation_errors"`
}

type ValidationError struct {
	FieldKey     string `json:"field_key"`
	ErrorMessage string `json:"error_message"`
}

func getErrorStruct(code ErrorCode) *ErrorStruct {
	errorStruct := &ErrorStruct{
		ErrorCode:    UnknownErrorCode,
		ErrorMessage: UnknownErrorMessage,
	}

	switch code {
	case UserAlreadyExistsCode:
		errorStruct.ErrorCode = UserAlreadyExistsCode
		errorStruct.ErrorMessage = UserAlreadyExistsMessage
	case UserNotFoundCode:
		errorStruct.ErrorCode = UserNotFoundCode
		errorStruct.ErrorMessage = UserNotFoundMessage
	case UserRefreshTokenCookieNotFoundCode:
		errorStruct.ErrorCode = UserRefreshTokenCookieNotFoundCode
		errorStruct.ErrorMessage = UserRefreshTokenCookieNotFoundMessage
	case UserRefreshTokenExpiredCode:
		errorStruct.ErrorCode = UserRefreshTokenExpiredCode
		errorStruct.ErrorMessage = UserRefreshTokenExpiredMessage
	}

	return errorStruct
}
