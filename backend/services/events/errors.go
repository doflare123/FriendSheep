package events

import "errors"

var (
	ErrEventNotFound         = errors.New("событие не найдено")
	ErrPermissionDenied      = errors.New("недостаточно прав")
	ErrNotGroupMember        = errors.New("вы не состоите в группе")
	ErrEventFull             = errors.New("событие заполнено")
	ErrAlreadyJoined         = errors.New("вы уже присоединились к событию")
	ErrNotJoined             = errors.New("вы не присоединялись к событию")
	ErrCreatorCantLeave      = errors.New("создатель не может покинуть событие")
	ErrInvalidEventUpdate    = errors.New("некорректное обновление события")
	ErrInvalidGenres         = errors.New("некорректные жанры")
	ErrEventTypeNotFound     = errors.New("тип события не найден")
	ErrEventLocationNotFound = errors.New("формат события не найден")
	ErrAgeLimitNotFound      = errors.New("возрастное ограничение не найдено")
	ErrEventAlreadyStarted   = errors.New("событие уже началось")
)
