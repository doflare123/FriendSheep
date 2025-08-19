package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
)

func DeleteAccount(email string) error {
	database := db.GetDB()

	var user models.User
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}
	if err := database.Delete(&user).Error; err != nil {
		return err
	}

	return nil
}
