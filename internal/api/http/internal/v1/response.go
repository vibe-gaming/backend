package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func errorResponse(c *gin.Context, code ErrorCode) {
	c.AbortWithStatusJSON(http.StatusBadRequest, getErrorStruct(code))
}

func validationErrorResponse(c *gin.Context, err error) {
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		out := make([]ValidationError, len(verr))
		for i, ferr := range verr {
			out[i] = ValidationError{ferr.Field(), msgForTag(ferr.Tag(), ferr.Param())}
		}
		response := ValidationErrorStruct{
			ErrorCode:    6000,
			ErrorMessage: "Validation error",
		}
		response.Errors = out
		c.JSON(http.StatusBadRequest, response)
	}
}

func msgForTag(tag string, value string) string {
	switch tag {
	case "required":
		return "Это поле обязательное к заполнению"
	case "email":
		return "Неверный формат почты"
	case "number":
		return "Поле должно иметь числовой формат"
	case "min":
		return fmt.Sprintf("Минимальное количество символов в поле - %v", value)
	case "max":
		return fmt.Sprintf("Максимальное количество символов в поле - %v", value)
	case "phonenumber":
		return "Номер должен начинаться с 7 и иметь 11 символов"
	}
	return tag
}
