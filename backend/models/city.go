package models

type City struct {
	ID        uint     `gorm:"primaryKey;autoIncrement"`
	Name      string   `gorm:"not null"`
	Region    string   `gorm:"not null"`
	FIASID    string   `gorm:"column:fias_id;type:varchar(36);uniqueIndex;not null"`
	Latitude  *float64 `gorm:"type:double precision"`
	Longitude *float64 `gorm:"type:double precision"`
}
