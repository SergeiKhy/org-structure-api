package logger

import (
	"context"
	"log/slog"
	"os"
)

var log *slog.Logger

// Init инициализирует логгер с нужным уровнем
func Init(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	log = slog.New(handler)
}

// Get возвращает глобальный логгер
func Get() *slog.Logger {
	return log
}

// Info логирует информационное сообщение
func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

// Error логирует ошибку
func Error(msg string, args ...any) {
	log.Error(msg, args...)
}

// Debug логирует отладочное сообщение
func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

// Warn логирует предупреждение
func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

// Logger для контекста запроса
type RequestLogger struct {
	logger *slog.Logger
}

// NewRequestLogger создаёт логгер для запроса
func NewRequestLogger() *RequestLogger {
	return &RequestLogger{
		logger: log,
	}
}

// LogRequest логирует HTTP запрос
func (rl *RequestLogger) LogRequest(ctx context.Context, method, path string, statusCode int, duration string) {
	rl.logger.InfoContext(ctx, "http request",
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", statusCode),
		slog.String("duration", duration),
	)
}

// LogError логирует ошибку запроса
func (rl *RequestLogger) LogError(ctx context.Context, method, path string, err error) {
	rl.logger.ErrorContext(ctx, "http request error",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("error", err.Error()),
	)
}
