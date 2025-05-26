package models

type SessionTypeAuth string

const (
	SessionTypeRegister      SessionTypeAuth = "register"
	SessionTypeResetPassword SessionTypeAuth = "reset_password"
)

type SessionAuth struct {
	Code       string      // 6-значный код подтверждения
	IsVerified bool        // подтверждена ли сессия
	Type       SessionType // тип сессии: регистрация или восстановление
}
