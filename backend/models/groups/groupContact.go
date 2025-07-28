package groups

// GroupContact представляет контактную информацию (социальную сеть) для группы.
type GroupContact struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name    string `json:"name" gorm:"not null"`
	Link    string `json:"link" gorm:"not null"`
	GroupID uint   `json:"groupId"`
}
