package register

import "context"

// UserBootstrap содержит данные, необходимые для создания зарегистрированного
// пользователя. Модели хранения остаются за границей RegistrationStore.
type UserBootstrap struct {
	Name           string
	HashedPassword string
	Email          string
	Username       string
}

// RegisteredUser содержит данные пользователя, нужные сценарию регистрации
// после сохранения учётной записи и начальных настроек.
type RegisteredUser struct {
	ID       uint
	Name     string
	Email    string
	Username string
	Image    string
}

// RegistrationStore описывает операции постоянного хранения для регистрации.
// CreateUser должен атомарно создать пользователя и все обязательные начальные строки.
type RegistrationStore interface {
	CreateUser(ctx context.Context, input UserBootstrap) (RegisteredUser, error)
	ChangePasswordByEmail(ctx context.Context, email, hashedPassword string) (uint, error)
}
