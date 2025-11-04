package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger() (*zap.Logger, error) {
	// Настройка ротации логов
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,   // мегабайты
		MaxBackups: 30,   // количество файлов
		MaxAge:     30,   // дни хранения
		Compress:   true, // gzip старых логов
	})

	consoleWriter := zapcore.AddSync(os.Stdout)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format(time.RFC3339))
	}
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	consoleEncoderCfg := zap.NewDevelopmentEncoderConfig()
	consoleEncoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	consoleEncoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, fileWriter, zap.DebugLevel),
		zapcore.NewCore(consoleEncoder, consoleWriter, zap.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger, nil
}
