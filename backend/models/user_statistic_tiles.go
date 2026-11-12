package models

import (
	"gorm.io/gorm"
)

// UserStatisticTiles хранит упорядоченный выбор плиток статистики пользователя.
type UserStatisticTiles struct {
	UserID       uint     `gorm:"primaryKey;autoIncrement:false"`
	SelectedKeys []string `gorm:"serializer:json;type:jsonb;not null"`
	User         User     `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserStatisticTiles) TableName() string { return "user_statistic_tiles" }

// BeforeCreate отличает отсутствующий выбор от явно сохранённого пустого списка.
func (tiles *UserStatisticTiles) BeforeCreate(_ *gorm.DB) error {
	if tiles.SelectedKeys == nil {
		tiles.SelectedKeys = DefaultUserStatisticTileKeys()
	}
	return nil
}

func DefaultUserStatisticTileKeys() []string {
	return []string{"films", "games", "tableGames", "completedEvents"}
}
