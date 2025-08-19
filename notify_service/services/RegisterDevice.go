package services

import (
	"context"
	"encoding/json"
	"fmt"
	"notify_service/db"
	"notify_service/models"
	"time"

	"gorm.io/gorm"
)

type DeviceRegistrationRequest struct {
	DeviceToken string                 `json:"device_token" validate:"required"`
	Platform    string                 `json:"platform" validate:"required,oneof=web android ios"`
	DeviceInfo  map[string]interface{} `json:"device_info,omitempty"`
}

// регистрирует новое устройство или обновляет существующее
func RegisterDevice(email string, req DeviceRegistrationRequest) error {
	if email == "" {
		return fmt.Errorf("не передан jwt")
	}
	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		return fmt.Errorf("failed to find user: %v", err)
	}

	userID := user.ID

	deviceInfoJSON := ""
	if req.DeviceInfo != nil {
		infoBytes, err := json.Marshal(req.DeviceInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal device info: %v", err)
		}
		deviceInfoJSON = string(infoBytes)
	}

	var existingDevice models.DeviceUser
	err := db.GetDB().Where("user_id = ? AND device_token = ?", userID, req.DeviceToken).First(&existingDevice).Error

	if err == gorm.ErrRecordNotFound {
		// Создаем новое устройство
		device := models.DeviceUser{
			UserID:      &userID,
			DeviceToken: &req.DeviceToken,
			Platform:    &req.Platform,
			DeviceInfo:  &deviceInfoJSON,
			IsActive:    boolPtr(true),
		}

		return db.GetDB().Create(&device).Error
	} else if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	updates := map[string]interface{}{
		"platform":    req.Platform,
		"device_info": deviceInfoJSON,
		"is_active":   true,
		"last_used":   time.Now(),
	}

	return db.GetDB().Model(&existingDevice).Updates(updates).Error
}

// отключает устройство
func UnregisterDevice(userID uint, deviceToken string) error {
	return db.GetDB().Model(&models.DeviceUser{}).
		Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Update("is_active", false).Error
}

// получает все активные устройства пользователя
func GetUserDevices(userID uint) ([]models.DeviceUser, error) {
	var devices []models.DeviceUser
	err := db.GetDB().Where("user_id = ? AND is_active = ?", userID, true).Find(&devices).Error
	return devices, err
}

// получает токены всех активных устройств пользователя
func GetActiveDeviceTokens(userID uint) ([]string, error) {
	var tokens []string
	err := db.GetDB().Model(&models.DeviceUser{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Pluck("device_token", &tokens).Error
	return tokens, err
}

// удаляет устройства, которые не использовались более N дней
func CleanupInactiveDevices(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return db.GetDB().Where("last_used < ? OR last_used IS NULL", cutoffDate).
		Delete(&models.DeviceUser{}).Error
}

// отправляет уведомление всем устройствам пользователя
func SendNotificationToUser(ctx context.Context, userID uint, title, body string, data map[string]string) error {
	tokens, err := GetActiveDeviceTokens(userID)
	if err != nil {
		return fmt.Errorf("failed to get user devices: %v", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	if len(tokens) == 1 {
		_, err = db.SendNotification(ctx, tokens[0], title, body, data)
	} else {
		_, err = db.SendMulticastNotification(ctx, tokens, title, body, data)
	}

	return err
}

func boolPtr(b bool) *bool {
	return &b
}
