package events

type EventLocation struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"unique"`
}
