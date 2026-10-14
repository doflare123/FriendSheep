package events

import (
	"fmt"

	"friendship/repository"
)

// Общая блокировка всех event outbox берётся до блокировок доменных строк и выдачи
// sequence. PostgreSQL удерживает её до завершения внешней транзакции, поэтому
// следующий издатель не обгоняет незавершённую публикацию по порядку commit.
// Ключ зарезервирован для event outbox; sequence должны сохранять CACHE 1.
func lockEventOutboxPublication(repo repository.PostgresRepository) error {
	if repo.Model(nil).Dialector.Name() != "postgres" {
		return nil
	}
	if err := repo.Exec("SELECT pg_advisory_xact_lock(?, ?)", int32(1179862864), int32(1)).Error; err != nil {
		return fmt.Errorf("ошибка блокировки публикации event outbox: %w", err)
	}
	return nil
}
