package logger

import "go.uber.org/zap"

type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Panic(msg string, fields ...interface{})
}

type ZapLogger struct {
	zap *zap.SugaredLogger
}

func NewZapLogger() (*ZapLogger, error) {
	base, err := NewLogger()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{zap: base.Sugar()}, nil
}

func (l *ZapLogger) Info(msg string, fields ...interface{})  { l.zap.Infow(msg, fields...) }
func (l *ZapLogger) Error(msg string, fields ...interface{}) { l.zap.Errorw(msg, fields...) }
func (l *ZapLogger) Debug(msg string, fields ...interface{}) { l.zap.Debugw(msg, fields...) }
func (l *ZapLogger) Warn(msg string, fields ...interface{})  { l.zap.Warnw(msg, fields...) }
func (l *ZapLogger) Fatal(msg string, fields ...interface{}) { l.zap.Fatalw(msg, fields...) }
func (l *ZapLogger) Panic(msg string, fields ...interface{}) { l.zap.Panicw(msg, fields...) }

func (l *ZapLogger) Zap() *zap.Logger {
	return l.zap.Desugar()
}
