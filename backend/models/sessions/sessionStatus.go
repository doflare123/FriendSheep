package sessions

import "time"

type Status struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Status    string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
