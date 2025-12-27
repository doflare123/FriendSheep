package events

type AgeLimit struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"not null;unique;size:50" json:"name"`
}
