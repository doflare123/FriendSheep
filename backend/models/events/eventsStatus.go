package events

const (
	StatusRecruitment = "Набор"
	StatusActive      = "В процессе"
	StatusCompleted   = "Завершена"
)

type Status struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null"`
}
