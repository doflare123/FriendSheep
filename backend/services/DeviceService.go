package services

import (
	"errors"
	"fmt"
	"friendship/db"
	"friendship/models"
	"time"

	"gorm.io/gorm"
)

type RegisterDeviceTokenRequest struct {
	DeviceToken string `json:"device_token" binding:"required"`
	Platform    string `json:"platform" binding:"required,oneof=ios android web"`
	DeviceInfo  string `json:"device_info"`
}

type DeviceTokenResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	DeviceToken string    `json:"device_token"`
	Platform    string    `json:"platform"`
	IsActive    bool      `json:"is_active"`
	LastUsed    time.Time `json:"last_used"`
	CreatedAt   time.Time `json:"created_at"`
}

// регистрирует новый токен или обновляет существующий
func RegisterOrUpdateDeviceToken(userID uint, req RegisterDeviceTokenRequest) (*DeviceTokenResponse, error) {
	database := db.GetDB()

	// Проверяем, существует ли пользователь
	var user models.User
	if err := database.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка при проверке пользователя: %v", err)
	}

	var existingDevice models.DeviceUser
	err := database.Where("device_token = ?", req.DeviceToken).First(&existingDevice).Error

	now := time.Now()
	isActive := true

	if err == nil {
		existingDevice.UserID = &userID
		existingDevice.Platform = &req.Platform
		existingDevice.IsActive = &isActive
		existingDevice.LastUsed = &now

		if req.DeviceInfo != "" {
			existingDevice.DeviceInfo = &req.DeviceInfo
		}

		if err := database.Save(&existingDevice).Error; err != nil {
			return nil, fmt.Errorf("ошибка при обновлении токена: %v", err)
		}

		return deviceToResponse(&existingDevice), nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("ошибка при проверке токена: %v", err)
	}

	newDevice := models.DeviceUser{
		UserID:      &userID,
		DeviceToken: &req.DeviceToken,
		Platform:    &req.Platform,
		IsActive:    &isActive,
		LastUsed:    &now,
	}

	if req.DeviceInfo != "" {
		newDevice.DeviceInfo = &req.DeviceInfo
	}

	if err := database.Create(&newDevice).Error; err != nil {
		return nil, fmt.Errorf("ошибка при создании токена: %v", err)
	}

	return deviceToResponse(&newDevice), nil
}

// получает все устройства пользователя
func GetUserDevices(userID uint) ([]DeviceTokenResponse, error) {
	database := db.GetDB()

	var devices []models.DeviceUser
	if err := database.Where("user_id = ?", userID).
		Order("last_used DESC").
		Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении устройств: %v", err)
	}

	responses := make([]DeviceTokenResponse, 0, len(devices))
	for _, device := range devices {
		responses = append(responses, *deviceToResponse(&device))
	}

	return responses, nil
}

// получает все активные токены пользователя
func GetActiveDeviceTokens(userID uint) ([]string, error) {
	database := db.GetDB()

	var devices []models.DeviceUser
	isActive := true
	if err := database.Where("user_id = ? AND is_active = ?", userID, isActive).
		Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении активных токенов: %v", err)
	}

	tokens := make([]string, 0, len(devices))
	for _, device := range devices {
		if device.DeviceToken != nil {
			tokens = append(tokens, *device.DeviceToken)
		}
	}

	return tokens, nil
}

// деактивирует токен устройства
func DeactivateDeviceToken(userID uint, deviceToken string) error {
	database := db.GetDB()

	isActive := false
	result := database.Model(&models.DeviceUser{}).
		Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Update("is_active", isActive)

	if result.Error != nil {
		return fmt.Errorf("ошибка при деактивации токена: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("токен не найден")
	}

	return nil
}

// удаляет токен устройства
func DeleteDeviceToken(userID uint, deviceToken string) error {
	database := db.GetDB()

	result := database.Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Delete(&models.DeviceUser{})

	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении токена: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("токен не найден")
	}

	return nil
}

// помечает токен как невалидный (для обработки ошибок FCM)
func InvalidateDeviceToken(deviceToken string) error {
	database := db.GetDB()

	isActive := false
	result := database.Model(&models.DeviceUser{}).
		Where("device_token = ?", deviceToken).
		Update("is_active", isActive)

	if result.Error != nil {
		return fmt.Errorf("ошибка при инвалидации токена: %v", result.Error)
	}

	return nil
}

// преобразует модель в ответ
func deviceToResponse(device *models.DeviceUser) *DeviceTokenResponse {
	response := &DeviceTokenResponse{}

	if device.ID != nil {
		response.ID = *device.ID
	}
	if device.UserID != nil {
		response.UserID = *device.UserID
	}
	if device.DeviceToken != nil {
		response.DeviceToken = *device.DeviceToken
	}
	if device.Platform != nil {
		response.Platform = *device.Platform
	}
	if device.IsActive != nil {
		response.IsActive = *device.IsActive
	}
	if device.LastUsed != nil {
		response.LastUsed = *device.LastUsed
	}
	response.CreatedAt = device.CreatedAt

	return response
}
