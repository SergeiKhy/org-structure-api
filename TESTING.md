# Руководство по тестированию

## Быстрая проверка

### 1. Unit тесты

```bash
cd /home/sergei/testTask
go test ./... -v
```

**Ожидаемый результат:**
```
ok  	github.com/SergeiKhy/org-structure-api/internal/config	0.002s
ok  	github.com/SergeiKhy/org-structure-api/internal/handler	0.005s
ok  	github.com/SergeiKhy/org-structure-api/internal/logger	0.002s
ok  	github.com/SergeiKhy/org-structure-api/internal/model	0.002s
ok  	github.com/SergeiKhy/org-structure-api/internal/repository	0.003s
ok  	github.com/SergeiKhy/org-structure-api/internal/service	0.004s
```

### 2. Сборка приложения

```bash
go build ./cmd/api
```

**Ожидаемый результат:** без ошибок, создаётся файл `api`

---

## Интеграционные тесты (требуют Docker)

### Подготовка

**Первый запуск** (скачивание образа PostgreSQL ~500MB):

```bash
docker pull postgres:15-alpine
```

**Проверка что Docker работает:**

```bash
docker ps
# Должен показать список контейнеров (может быть пустым)
```

### Запуск интеграционных тестов

```bash
# Все интеграционные тесты (~2-3 минуты при первом запуске)
go test ./... -tags=integration -v

# Только repository layer (~1 минута)
go test ./internal/repository/... -tags=integration -v

# Только service layer (~1 минута)
go test ./internal/service/... -tags=integration -v
```

### Что происходит при запуске:

1. **testcontainers-go** запускает контейнер PostgreSQL 15
2. Создаются таблицы через GORM AutoMigrate
3. Выполняются тесты
4. Контейнер удаляется

### Ожидаемый вывод:

```
=== RUN   TestRepository_CreateDepartment_Integration
--- PASS: TestRepository_CreateDepartment_Integration (2.34s)
=== RUN   TestRepository_CheckUniqueName_Integration
--- PASS: TestRepository_CheckUniqueName_Integration (0.45s)
=== RUN   TestService_CreateDepartment_Integration
--- PASS: TestService_CreateDepartment_Integration (1.89s)
PASS
ok  	github.com/SergeiKhy/org-structure-api/internal/repository	5.67s
```

---

## Ручное тестирование API

### Запуск через docker-compose

```bash
cd /home/sergei/testTask

# Сборка и запуск
docker-compose up --build

# Или в фоновом режиме
docker-compose up -d --build
```

### Проверка что работает:

```bash
# Проверка контейнеров
docker-compose ps

# Логи приложения
docker-compose logs app
```

### Тестирование endpoints:

```bash
# 1. Создать подразделение
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Engineering"}'

# 2. Получить подразделение
curl http://localhost:8080/departments/1

# 3. Создать сотрудника
curl -X POST http://localhost:8080/departments/1/employees/ \
  -H "Content-Type: application/json" \
  -d '{"full_name": "John Doe", "position": "Developer"}'
```

### Остановка:

```bash
docker-compose down
```

---

## Покрытие кода

### Unit тесты:

```bash
go test ./... -cover
```

### Интеграционные тесты:

```bash
go test ./... -tags=integration -coverprofile=integration-coverage.out
go tool cover -html=integration-coverage.out  # Откроет браузер
```

---

## Возможные проблемы

### 1. Docker не запущен

**Linux:**
```bash
sudo systemctl start docker
```

**macOS:**
```bash
open -a Docker
```

**Windows:**
Запустите Docker Desktop

### 2. Нет прав на Docker socket (Linux)

```bash
sudo usermod -aG docker $USER
# Перезайдите в систему
```

### 3. Тесты "зависают" при первом запуске

Это нормально — скачивается образ PostgreSQL (~500MB).

**Решение:** предварительно скачайте образ:
```bash
docker pull postgres:15-alpine
```

### 4. Ошибка "connection refused"

Проверьте что БД запущена:
```bash
docker-compose ps
```

Перезапустите:
```bash
docker-compose down && docker-compose up -d
```

### 5. Порт 8080 занят

Измените порт в `.env`:
```
SERVER_PORT=8081
```
