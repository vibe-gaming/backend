package validator

import (
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterGinValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		err := v.RegisterValidation("phonenumber", phoneNumberValidator)
		if err != nil {
			log.Fatal("register phonenumber validator failed")
		}
	}
}

var phoneNumberValidator validator.Func = func(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()
	pattern := `^7\d{10}$`
	matched, err := regexp.MatchString(pattern, phoneNumber)
	if err != nil {
		return false
	}
	return matched
}
