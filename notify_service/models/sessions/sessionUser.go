package sessions

import "time"

type SessionUser struct {
	ID        uint `gorm:"primaryKey"`
	SessionID uint
	Session   Session   `gorm:"foreignKey:SessionID"`
	UserID    uint      `gorm:"not null;index"`
	JoinedAt  time.Time `gorm:"autoCreateTime"`
}
