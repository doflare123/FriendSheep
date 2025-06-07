package services

import "github.com/go-playground/validator/v10"

var Validate *validator.Validate

func InitValidator(v *validator.Validate) {
	Validate = v
}

func GetValidator() *validator.Validate {
	return Validate
}
