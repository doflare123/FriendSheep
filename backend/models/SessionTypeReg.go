package models

//redis

type SessionTypeReg string

const (
	SessionTypeRegister      SessionTypeReg = "register"
	SessionTypeResetPassword SessionTypeReg = "reset_password"
)

type SessionRegResponse struct {
	SessionID string `json:"session_id"`
}

type SessionReg struct {
	Code        string
	is_verified bool
	Type        SessionType
	attempts    uint8
}
