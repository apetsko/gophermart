package logging

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogEntry interface {
	Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{})
	Panic(v interface{}, stack []byte)
}

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(level zapcore.Level) (*Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.StacktraceKey = ""
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level.SetLevel(level)

	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1)) // Добавляем AddCaller и сдвиг
	if err != nil {
		return nil, err
	}
	return &Logger{logger.Sugar()}, nil
}

func (l *Logger) Close() error {
	return l.Sync()
}

func (l *Logger) Debug(message string, keysAndValues ...interface{}) {
	l.Debugw(message, keysAndValues...)
}

func (l *Logger) Info(message string, keysAndValues ...interface{}) {
	l.Infow(message, keysAndValues...)
}

func (l *Logger) Error(message string, keysAndValues ...interface{}) {
	l.Errorw(message, keysAndValues...)
}

func (l *Logger) Fatal(message string, keysAndValues ...interface{}) {
	l.Fatalw(message, keysAndValues...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Infof(format, v...)
}
