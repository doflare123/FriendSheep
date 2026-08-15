package migrations_test

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"notify_service/migrations"
)

func TestEmbeddedMigrationHistoryIsOrderedAndServiceOwned(t *testing.T) {
	t.Parallel()

	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		t.Fatalf("не удалось прочитать встроенные миграции: %v", err)
	}

	namePattern := regexp.MustCompile(`^(\d{6})_[a-z0-9_]+\.sql$`)
	var names []string
	var previousVersion int64
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		matches := namePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			t.Fatalf("некорректное имя файла миграции %q", entry.Name())
		}
		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			t.Fatalf("не удалось разобрать версию миграции %q: %v", entry.Name(), err)
		}
		if version <= previousVersion {
			t.Fatalf("версия миграции %d не новее предыдущей версии %d", version, previousVersion)
		}
		previousVersion = version
		names = append(names, entry.Name())

		body, err := fs.ReadFile(migrations.Files, entry.Name())
		if err != nil {
			t.Fatalf("не удалось прочитать миграцию %q: %v", entry.Name(), err)
		}
		lowerBody := strings.ToLower(string(body))
		if !strings.Contains(lowerBody, "notify_service") {
			t.Errorf("миграция %q не использует собственную схему notify_service", entry.Name())
		}
		for _, forbiddenTable := range []string{"public.events", "public.sessions", "public.users", "public.groups"} {
			if strings.Contains(lowerBody, forbiddenTable) {
				t.Errorf("миграция %q ссылается на таблицу монолита %q", entry.Name(), forbiddenTable)
			}
		}
	}

	if len(names) == 0 {
		t.Fatal("не найдены встроенные миграции сервиса уведомлений")
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("встроенные миграции не упорядочены лексикографически: %v", names)
	}
}

func TestPublishedMigrationChecksumsRemainImmutable(t *testing.T) {
	t.Parallel()

	publishedChecksums := map[string]string{
		"000001_initial_schema.sql": "ed2163a41c532626bad6c2d225605ed133e2cf574da7792a9adff32ad5782460",
	}

	for name, wantChecksum := range publishedChecksums {
		body, err := fs.ReadFile(migrations.Files, name)
		if err != nil {
			t.Fatalf("не удалось прочитать опубликованную миграцию %q: %v", name, err)
		}
		gotChecksum := fmt.Sprintf("%x", sha256.Sum256(body))
		if gotChecksum != wantChecksum {
			t.Errorf("контрольная сумма опубликованной миграции %q изменилась: получено %s, ожидалось %s", name, gotChecksum, wantChecksum)
		}
	}
}
