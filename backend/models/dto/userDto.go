package dto

import "time"

type UserDto struct {
	Id           int64     `json:"id"`
	Username     string    `json:"username"`
	Us           string    `json:"us"`
	Image        string    `json:"image"`
	DataRegister time.Time `json:"data_register"`
	Enterprise   bool      `json:"enterprise"`
	VerifiedUser bool      `json:"verified"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Telegram     bool      `json:"link_tg"`
}
