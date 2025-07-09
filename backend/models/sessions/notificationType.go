package sessions

type NotificationType struct {
	ID     uint   `gorm:"primaryKey"`
	Code   string `gorm:"unique;not null"` // "1d", "6h", etc.
	Label  string `gorm:"not null"`        // "За 1 день", "За 6 часов"
	Offset int64  `gorm:"not null"`        // в секундах: -86400, -21600
}
