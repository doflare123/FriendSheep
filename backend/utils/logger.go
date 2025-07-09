package utils

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var Log *log.Logger

func Init() {
	logDir := "logs"
	os.MkdirAll(logDir, os.ModePerm)

	// Имя файла по дате
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Не удалось создать лог-файл:", err)
	}

	// Лог в файл + консоль
	multiWriter := io.MultiWriter(os.Stdout, file)
	Log = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

	// Удаление старых логов
	cleanupOldLogs(logDir, 12)
}

func cleanupOldLogs(logDir string, maxAgeDays int) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		Log.Println("Ошибка чтения папки логов:", err)
		return
	}

	threshold := time.Now().AddDate(0, 0, -maxAgeDays)

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(threshold) {
			os.Remove(filepath.Join(logDir, file.Name()))
		}
	}
}
