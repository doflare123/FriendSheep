package services

import (
	"friendship/db"
	"friendship/models"
)

func FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
