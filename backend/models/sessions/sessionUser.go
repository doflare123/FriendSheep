package sessions

import "time"

type SessionUser struct {
	ID        uint      `gorm:"primaryKey"`
	SessionID uint      `gorm:"not null;index"`
	UserID    uint      `gorm:"not null;index"`
	JoinedAt  time.Time `gorm:"autoCreateTime"`
}
