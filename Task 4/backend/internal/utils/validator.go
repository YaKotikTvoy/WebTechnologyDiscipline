package utils

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

func init() {
	Validate.RegisterValidation("phone", validatePhone)
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if len(phone) < 10 || len(phone) > 20 {
		return false
	}
	return true
}
