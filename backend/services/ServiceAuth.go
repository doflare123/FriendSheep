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

func FindUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := db.GetDB().Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUserByUs(us string) (*uint, error) {
	var user models.User
	if err := db.GetDB().Where("us = ?", us).First(&user).Error; err != nil {
		return nil, err
	}
	return &user.ID, nil
}
