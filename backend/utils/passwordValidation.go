package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func PasswordValidation(fl validator.FieldLevel) bool {
	pass := fl.Field().String()
	if len(pass) < 10 {
		return false
	}
	var hasLower, hasUpper, hasSpecial bool
	for _, c := range pass {
		switch {
		case 'a' <= c && c <= 'z':
			hasLower = true
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case (c >= 33 && c <= 47) || (c >= 58 && c <= 64) || (c >= 91 && c <= 96) || (c >= 123 && c <= 126):
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasSpecial
}

func ValidateNameTag(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	if name == "" {
		return false
	}

	// Паттерн для разрешенных символов: буквы (латиница и кириллица), цифры, пробелы, дефисы, подчеркивания
	allowedPattern := regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9\s\-_]+$`)

	return allowedPattern.MatchString(name)
}
