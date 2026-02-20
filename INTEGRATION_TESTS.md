# Интеграционные тесты

## Обзор

Интеграционные тесты в проекте используют **testcontainers-go** для запуска реального PostgreSQL в Docker контейнере.

## Запуск тестов

### Все интеграционные тесты

```bash
go test ./... -tags=integration -v
```

### Тесты конкретного пакета

```bash
# Repository layer
go test ./internal/repository/... -tags=integration -v

# Service layer  
go test ./internal/service/... -tags=integration -v
```

### С покрытием кода

```bash
go test ./... -tags=integration -cover
```

### С выводом покрытия в HTML

```bash
go test ./... -tags=integration -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Требования

- Docker (для запуска контейнеров PostgreSQL)
- Go 1.24+

## Как это работает

1. Перед каждым тестом запускается контейнер PostgreSQL 15
2. Создаются таблицы через GORM AutoMigrate
3. Выполняется тест
4. Контейнер удаляется

## Пример теста

```go
//go:build integration

func TestRepository_CreateDepartment_Integration(t *testing.T) {
    // Запускаем контейнер
    pgContainer, db, ctx := setupTestContainer(t)
    defer pgContainer.Terminate(ctx)

    repo := NewRepository(db)

    // Тестируем
    dept := &model.Department{Name: "Engineering"}
    err := repo.CreateDepartment(dept)
    
    if err != nil {
        t.Fatalf("ошибка: %v", err)
    }
}
```

## Разделение тестов

- **Unit тесты** (без тегов) — быстрые, с моками
- **Интеграционные** (`-tags=integration`) — медленнее, с реальной БД

## CI/CD

В GitHub Actions используйте:

```yaml
- name: Run integration tests
  run: go test ./... -tags=integration -v
```

Docker будет доступен автоматически.
