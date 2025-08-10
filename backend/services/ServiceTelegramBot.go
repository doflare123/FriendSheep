package services

import (
	"errors"
	"friendship/db"
	"friendship/models"
)

type TelegramNotifyInput struct {
	Us         *string `json:"us"`
	TelegramID *string `json:"telegram_id"`
}

func (s *TelegramNotifyInput) SubscriptionsTelegramNotify() error {
	var user models.User
	if err := db.GetDB().Where("us = ?", s.Us).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}
	if user.TelegramID != nil {
		return errors.New("пользователь уже подписан")
	}
	user.TelegramID = s.TelegramID
	if err := db.GetDB().Save(&user).Error; err != nil {
		return errors.New("ошибка при сохранении пользователя")
	}

	return nil
}

func (s *TelegramNotifyInput) UnSubscriptionsTelegramNotify() error {
	if s.TelegramID == nil {
		return errors.New("не указан TelegramID")
	}
	var user models.User
	if err := db.GetDB().Where("telegram_id = ?", s.TelegramID).First(&user).Error; err != nil {
		return errors.New("пользователь не найден")
	}
	user.TelegramID = nil
	if err := db.GetDB().Save(&user).Error; err != nil {
		return errors.New("ошибка при сохранении пользователя")
	}

	return nil
}
