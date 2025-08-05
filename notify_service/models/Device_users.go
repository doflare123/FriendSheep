package models

type Device_Users struct {
	ID          uint     `gorm:"primaryKey;autoIncrement"`
	UserID      *uint    `gorm:"not null;index"`
	User        User     `gorm:"foreignKey:UserID;references:ID"`
	DeviceToken string   `json:"device_token" gorm:"not null;index"`
	Platform    Platform `gorm:"foreignKey:PlatformID;references:ID"`
	PlatformID  uint     `json:"platform_id" gorm:"not null"`
}
