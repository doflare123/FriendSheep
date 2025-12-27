package events

import "time"

type Genre struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null;unique;size:100" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type EventGenre struct {
	ID      uint  `gorm:"primaryKey;autoIncrement"`
	EventID uint  `gorm:"not null;index" json:"eventId"`
	Event   Event `gorm:"foreignKey:EventID" json:"-"`
	GenreID uint  `gorm:"not null;index" json:"genreId"`
	Genre   Genre `gorm:"foreignKey:GenreID" json:"genre"`
}
