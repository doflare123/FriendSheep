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
	Code        string      // 6-значный код подтверждения
	is_verified bool        // подтверждена ли сессия
	Type        SessionType // тип сессии: регистрация или восстановление
}
