// utils/logger.go
package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	PANIC
)

type Logger struct {
	*log.Logger
	level LogLevel
}

var Log *Logger

func Init() {
	logDir := "logs"
	os.MkdirAll(logDir, os.ModePerm)

	// Создаем файлы для разных уровней логирования
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	errorFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+"_error.log")

	// Основной файл лога
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Не удалось создать лог-файл:", err)
	}

	// Файл для ошибок
	errorFileHandle, err := os.OpenFile(errorFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Не удалось создать файл ошибок:", err)
	}

	// Настраиваем MultiWriter для разных уровней
	var writer io.Writer

	// В production логируем только в файлы, в development - еще и в консоль
	if os.Getenv("GIN_MODE") == "release" {
		writer = file
	} else {
		writer = io.MultiWriter(os.Stdout, file)
	}

	baseLogger := log.New(writer, "", 0) // Убираем стандартные флаги, добавим свои

	Log = &Logger{
		Logger: baseLogger,
		level:  INFO, // Уровень по умолчанию
	}

	// Устанавливаем уровень логирования из переменной окружения
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		Log.SetLevel(logLevel)
	}

	// Настраиваем отдельный логгер для критических ошибок
	setupErrorLogger(errorFileHandle)

	// Удаление старых логов
	cleanupOldLogs(logDir, 30) // Увеличиваем до 30 дней

	Log.Info("Logger initialized successfully")
}

func (l *Logger) SetLevel(levelStr string) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		l.level = DEBUG
	case "INFO":
		l.level = INFO
	case "WARN":
		l.level = WARN
	case "ERROR":
		l.level = ERROR
	case "PANIC":
		l.level = PANIC
	default:
		l.level = INFO
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

func (l *Logger) formatMessage(level string, msg string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Получаем информацию о вызывающей функции
	_, file, line, ok := runtime.Caller(3)
	var caller string
	if ok {
		filename := filepath.Base(file)
		caller = fmt.Sprintf("%s:%d", filename, line)
	} else {
		caller = "unknown"
	}

	return fmt.Sprintf("[%s] %s [%s] %s", timestamp, level, caller, msg)
}

func (l *Logger) Debug(v ...interface{}) {
	if l.shouldLog(DEBUG) {
		msg := fmt.Sprint(v...)
		l.Output(0, l.formatMessage("DEBUG", msg))
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.shouldLog(DEBUG) {
		msg := fmt.Sprintf(format, v...)
		l.Output(0, l.formatMessage("DEBUG", msg))
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.shouldLog(INFO) {
		msg := fmt.Sprint(v...)
		l.Output(0, l.formatMessage("INFO", msg))
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.shouldLog(INFO) {
		msg := fmt.Sprintf(format, v...)
		l.Output(0, l.formatMessage("INFO", msg))
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if l.shouldLog(WARN) {
		msg := fmt.Sprint(v...)
		l.Output(0, l.formatMessage("WARN", msg))
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.shouldLog(WARN) {
		msg := fmt.Sprintf(format, v...)
		l.Output(0, l.formatMessage("WARN", msg))
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.shouldLog(ERROR) {
		msg := fmt.Sprint(v...)
		formatted := l.formatMessage("ERROR", msg)
		l.Output(0, formatted)

		// Дополнительно логируем в файл ошибок
		if errorLogger != nil {
			errorLogger.Output(0, formatted)
		}
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.shouldLog(ERROR) {
		msg := fmt.Sprintf(format, v...)
		formatted := l.formatMessage("ERROR", msg)
		l.Output(0, formatted)

		if errorLogger != nil {
			errorLogger.Output(0, formatted)
		}
	}
}

func (l *Logger) Panic(v ...interface{}) {
	msg := fmt.Sprint(v...)
	formatted := l.formatMessage("PANIC", msg)
	l.Output(0, formatted)

	if errorLogger != nil {
		errorLogger.Output(0, formatted)
	}

	panic(msg)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	formatted := l.formatMessage("PANIC", msg)
	l.Output(0, formatted)

	if errorLogger != nil {
		errorLogger.Output(0, formatted)
	}

	panic(msg)
}

// Для обратной совместимости с старым кодом
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Infof(format, v...)
}

func (l *Logger) Println(v ...interface{}) {
	l.Info(v...)
}

var errorLogger *log.Logger

func setupErrorLogger(errorFile *os.File) {
	var writer io.Writer
	if os.Getenv("GIN_MODE") == "release" {
		writer = errorFile
	} else {
		writer = io.MultiWriter(os.Stderr, errorFile)
	}
	errorLogger = log.New(writer, "", 0)
}

func cleanupOldLogs(logDir string, maxAgeDays int) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		if Log != nil {
			Log.Error("Ошибка чтения папки логов:", err)
		}
		return
	}

	threshold := time.Now().AddDate(0, 0, -maxAgeDays)
	cleaned := 0

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(threshold) {
			if err := os.Remove(filepath.Join(logDir, file.Name())); err == nil {
				cleaned++
			}
		}
	}

	if Log != nil && cleaned > 0 {
		Log.Infof("Cleaned up %d old log files", cleaned)
	}
}

// Дополнительные утилиты для логирования

func LogRequest(method, path, ip string, duration time.Duration, status int) {
	level := "INFO"
	if status >= 500 {
		level = "ERROR"
	} else if status >= 400 {
		level = "WARN"
	}

	Log.Printf("%s: %s %s [%d] %s from %s",
		level, method, path, status, duration, ip)
}

func LogError(err error, context string) {
	if err != nil {
		Log.Errorf("Error in %s: %v", context, err)
	}
}

func LogPanic(recovered interface{}, context string) {
	Log.Panicf("Panic in %s: %v", context, recovered)
}
