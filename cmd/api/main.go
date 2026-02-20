package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SergeiKhy/org-structure-api/internal/config"
	"github.com/SergeiKhy/org-structure-api/internal/handler"
	"github.com/SergeiKhy/org-structure-api/internal/logger"
	"github.com/SergeiKhy/org-structure-api/internal/repository"
	"github.com/SergeiKhy/org-structure-api/internal/service"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Инициализируем логгер
	logger.Init("info")
	log := logger.Get()

	log.Info("запуск приложения")

	cfg := config.Load()

	// DSN для GORM
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	// Подключение GORM
	log.Info("подключение к базе данных",
		slog.String("host", cfg.DBHost),
		slog.String("port", cfg.DBPort),
		slog.String("database", cfg.DBName))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error("ошибка подключения к базе данных",
			slog.String("error", err.Error()))
		return
	}

	log.Info("успешное подключение к базе данных")

	// Запуск миграции Goose
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("ошибка получения sql.DB",
			slog.String("error", err.Error()))
		return
	}

	log.Info("запуск миграций")
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Error("ошибка выполнения миграций",
			slog.String("error", err.Error()))
		return
	}

	log.Info("миграции успешно выполнены")

	// Инициализация слоев
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	hndl := handler.NewHandler(svc)

	// Создаём логгер запросов
	reqLogger := logger.NewRequestLogger()

	// Роутинг с логгированием
	http.HandleFunc("/departments/", func(w http.ResponseWriter, r *http.Request) {
		// Запоминаем время начала запроса
		start := time.Now()

		// Создаём контекст с ID запроса
		ctx := context.WithValue(r.Context(), "request_id", time.Now().UnixNano())
		r = r.WithContext(ctx)

		// Обёртка для перехвата статуса ответа
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		path := strings.TrimPrefix(r.URL.Path, "/departments/")
		parts := strings.Split(path, "/")

		// Если путь пустой или slash, создание подразделения
		if len(parts) == 0 || parts[0] == "" {
			if r.Method == http.MethodPost {
				hndl.CreateDepartment(rw, r)
			} else {
				hndl.WriteError(rw, http.StatusNotFound, "not found")
			}
			reqLogger.LogRequest(ctx, r.Method, r.URL.Path, rw.statusCode, time.Since(start).String())
			return
		}

		// Проверка на вложенный ресурс employees
		if len(parts) >= 2 && parts[1] == "employees" {
			if r.Method == http.MethodPost {
				hndl.CreateEmployee(rw, r)
			} else {
				hndl.WriteError(rw, http.StatusMethodNotAllowed, "method not allowed")
			}
			reqLogger.LogRequest(ctx, r.Method, r.URL.Path, rw.statusCode, time.Since(start).String())
			return
		}

		// Работа с конкретным департаментом (/departments/{id})
		switch r.Method {
		case http.MethodGet:
			hndl.GetDepartment(rw, r)
		case http.MethodPatch:
			hndl.UpdateDepartment(rw, r)
		case http.MethodDelete:
			hndl.DeleteDepartment(rw, r)
		default:
			hndl.WriteError(rw, http.StatusMethodNotAllowed, "method not allowed")
		}

		// Логируем запрос
		reqLogger.LogRequest(ctx, r.Method, r.URL.Path, rw.statusCode, time.Since(start).String())
	})

	log.Info("сервер запущен",
		slog.String("port", cfg.ServerPort),
		slog.String("environment", getEnv("ENVIRONMENT", "development")))

	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Error("ошибка запуска сервера",
			slog.String("error", err.Error()))
	}
}

// responseWriter обёртка для перехвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
