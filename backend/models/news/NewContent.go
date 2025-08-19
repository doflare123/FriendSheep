package news

type ContentNews struct {
	ID     int    `json:"id" gorm:"primaryKey;autoIncrement"`
	NewsID uint   `json:"news_id" gorm:"not null;index"`
	Text   string `json:"text" gorm:"not null" validate:"required"`
}
