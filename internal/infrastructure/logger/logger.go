package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

// Initialize настраивает и инициализирует zap-логгер с заданным уровнем
func Initialize(level string) error {
	var config zap.Config

	if level == "debug" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	var zapLevel zap.AtomicLevel
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	config.Level = zapLevel

	var err error
	Log, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	return nil
}

// InitializeNoop no‑op логгер для тестов (без вывода)
func InitializeNoop() {
	Log = zap.NewNop()
}

func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
}

// Sync сбрасывает буферы логгера
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
