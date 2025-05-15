package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func PasswordValidation(fl validator.FieldLevel) bool {
	const regex = `^(?=.*[a-z])(?=.*[A-Z])(?=.*[\W_]).{10,}$`
	return regexp.MustCompile(regex).MatchString(fl.Field().String())
}
