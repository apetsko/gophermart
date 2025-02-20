package logging

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	*zap.SugaredLogger
}

func NewZapLogger() (*ZapLogger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.StacktraceKey = ""
	config.Level.SetLevel(zap.DebugLevel)

	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1)) // Добавляем AddCaller и сдвиг
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger.Sugar()}, nil
}

func (l *ZapLogger) Close() error {
	return l.Sync()
}

func (l *ZapLogger) Info(message string, keysAndValues ...interface{}) {
	l.Infow(message, keysAndValues...)
}

func (l *ZapLogger) Error(message string, keysAndValues ...interface{}) {
	l.Errorw(message, keysAndValues...)
}

func (l *ZapLogger) Fatal(message string, keysAndValues ...interface{}) {
	l.Fatalw(message, keysAndValues...)
}
