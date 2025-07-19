package statsusers

type Genre struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}
