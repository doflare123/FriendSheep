package services

import (
	"notify_service/db"
	"notify_service/models"

	"gorm.io/gorm"
)

type RegisterDeviceInput struct {
	deviceToken  string
	deviceTypeID uint
}

func RegisterDevice(email string, input RegisterDeviceInput) error {
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return err
	}

	userID := user.ID

	var device models.Device_Users
	if err := db.GetDB().Where("email = ? AND device_token = ? AND platform_id = ?", email, input.deviceToken, input.deviceTypeID).First(&device).Error; err == gorm.ErrRecordNotFound {
		// Создаем новое устройство
		device := models.Device_Users{
			UserID:      &userID,
			DeviceToken: input.deviceToken,
			PlatformID:  input.deviceTypeID,
		}
		return db.GetDB().Create(&device).Error
	} else if err != nil {
		return err
	}

	return nil
}
