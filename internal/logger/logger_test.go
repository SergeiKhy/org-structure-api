package logger

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestInit проверяет инициализацию логгера
func TestInit(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"default level", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.level)
			log := Get()
			if log == nil {
				t.Error("ожидался логгер, получен nil")
			}
		})
	}
}

// TestInfo проверяет логирование информационного сообщения
func TestInfo(t *testing.T) {
	Init("info")
	// Просто проверяем что не паникует
	Info("тестовое сообщение", "key", "value")
}

// TestError проверяет логирование ошибки
func TestError(t *testing.T) {
	Init("info")
	Error("тестовая ошибка", "error", "something went wrong")
}

// TestDebug проверяет логирование отладочного сообщения
func TestDebug(t *testing.T) {
	Init("debug")
	Debug("отладочное сообщение", "debug", true)
}

// TestWarn проверяет логирование предупреждения
func TestWarn(t *testing.T) {
	Init("info")
	Warn("предупреждение", "code", 42)
}

// TestRequestLogger_LogRequest проверяет логирование запроса
func TestRequestLogger_LogRequest(t *testing.T) {
	Init("info")
	rl := NewRequestLogger()

	ctx := context.Background()
	// Проверяем что метод не паникует
	rl.LogRequest(ctx, "GET", "/departments/1", 200, "15ms")
}

// TestRequestLogger_LogError проверяет логирование ошибки запроса
func TestRequestLogger_LogError(t *testing.T) {
	Init("info")
	rl := NewRequestLogger()

	ctx := context.Background()
	err := errors.New("тестовая ошибка")
	// Проверяем что метод не паникует
	rl.LogError(ctx, "POST", "/departments/", err)
}

// TestRequestLogger_WithContext проверяет логгер с контекстом
func TestRequestLogger_WithContext(t *testing.T) {
	Init("info")
	rl := NewRequestLogger()

	ctx := context.WithValue(context.Background(), "request_id", time.Now().UnixNano())
	rl.LogRequest(ctx, "GET", "/departments", 200, "10ms")
}

// BenchmarkInit измеряет производительность инициализации
func BenchmarkInit(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Init("info")
	}
}

// BenchmarkInfo измеряет производительность логирования Info
func BenchmarkInfo(b *testing.B) {
	Init("info")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark message", "iteration", i)
	}
}

// BenchmarkLogRequest измеряет производительность логирования запроса
func BenchmarkLogRequest(b *testing.B) {
	Init("info")
	rl := NewRequestLogger()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.LogRequest(ctx, "GET", "/departments/1", 200, "5ms")
	}
}
