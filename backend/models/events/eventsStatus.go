package events

type Status struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null"`
}
