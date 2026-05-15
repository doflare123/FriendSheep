package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ValidationError(c *gin.Context, err error) {
	if errs, ok := err.(validator.ValidationErrors); ok {
		errorsMap := make(map[string]string)
		for _, e := range errs {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errorsMap[field] = "Поле обязательно для заполнения"
			case "min":
				errorsMap[field] = fmt.Sprintf("Минимальная длина %s символов", e.Param())
			case "max":
				errorsMap[field] = fmt.Sprintf("Максимальная длина %s символов", e.Param())
			case "email":
				errorsMap[field] = "Неверный формат email"
			case "oneof":
				errorsMap[field] = fmt.Sprintf("Допустимые значения: %s", e.Param())
			default:
				errorsMap[field] = "Некорректное значение"
			}
		}

		BadRequest(c, "Некорректные данные формы", WithFields(errorsMap))
		return
	}

	BadRequest(c, err.Error())
}
