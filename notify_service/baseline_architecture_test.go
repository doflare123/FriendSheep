package main

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLegacyMonolithCouplingArtifactsAreRemoved(t *testing.T) {
	t.Parallel()

	root := notifyServiceRoot(t)
	legacyPaths := []string{
		"schedulers",
		filepath.Join("models", "sessions"),
		filepath.Join("models", "users.go"),
		filepath.Join("models", "groups"),
		filepath.Join("utils", "jwt.go"),
		filepath.Join("middleware", "jwt.go"),
	}

	for _, relativePath := range legacyPaths {
		relativePath := relativePath
		t.Run(relativePath, func(t *testing.T) {
			_, err := os.Stat(filepath.Join(root, relativePath))
			if err == nil {
				t.Fatalf("устаревший артефакт со связью с монолитом всё ещё существует: %s", relativePath)
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("не удалось проверить %s: %v", relativePath, err)
			}
		})
	}
}

func TestCredentialFilesAreNotPresentInServiceTree(t *testing.T) {
	t.Parallel()

	root := notifyServiceRoot(t)
	secretPaths := []string{
		".env",
		"firebase-credentials.json",
		filepath.Join("db", "key.json"),
	}

	for _, relativePath := range secretPaths {
		if _, err := os.Stat(filepath.Join(root, relativePath)); err == nil {
			t.Errorf("файл с учётными данными не должен находиться в notify_service: %s", relativePath)
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("не удалось проверить %s: %v", relativePath, err)
		}
	}
}

func TestProductionCodeDoesNotReferenceDisabledLegacyRuntime(t *testing.T) {
	t.Parallel()

	root := notifyServiceRoot(t)
	forbidden := []string{
		"StartNotificationWorker",
		"notify_service/schedulers",
		"notify_service/models/sessions",
		"firebase.google.com",
		"jwt.NewWithClaims",
		"godotenv.Load",
	}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == ".cache" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, marker := range forbidden {
			if strings.Contains(string(contents), marker) {
				relativePath, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				t.Errorf("рабочий код %s ссылается на отключённый маркер устаревшего выполнения %q", relativePath, marker)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("не удалось проверить рабочий код: %v", err)
	}
}

func TestLegacyComposeKeepsNotifyServiceIsolated(t *testing.T) {
	t.Parallel()

	composePath := filepath.Join(filepath.Dir(notifyServiceRoot(t)), "docker", "docker-compose.override.yml")
	contents, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("не удалось прочитать устаревшую конфигурацию compose: %v", err)
	}
	serviceBlock := composeServiceBlock(t, string(contents), "notify_service")

	if !strings.Contains(serviceBlock, "NOTIFY_DATABASE_URL") {
		t.Fatal("блок notify_service в устаревшей конфигурации compose не содержит отдельный NOTIFY_DATABASE_URL")
	}
	for marker, explanation := range map[string]string{
		"../notify_service/.env": "монтирование старого .env",
		"env_file:":              "чтение старого env_file",
		"depends_on:":            "зависимость от контейнеров монолита или Telegram",
	} {
		if strings.Contains(serviceBlock, marker) {
			t.Errorf("блок notify_service в устаревшей конфигурации compose содержит запрещённое поведение: %s", explanation)
		}
	}
}

func composeServiceBlock(t *testing.T, contents, service string) string {
	t.Helper()

	header := "  " + service + ":"
	var lines []string
	inside := false
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == header {
			inside = true
			lines = append(lines, line)
			continue
		}
		if inside && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			break
		}
		if inside {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("не удалось разобрать устаревшую конфигурацию compose: %v", err)
	}
	if !inside {
		t.Fatalf("в устаревшей конфигурации compose отсутствует блок сервиса %q", service)
	}
	return strings.Join(lines, "\n")
}

func notifyServiceRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("не удалось определить путь к тестовому файлу")
	}
	return filepath.Dir(currentFile)
}
