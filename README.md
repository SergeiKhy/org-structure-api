# Organization Structure API

API для управления организационной структурой компании (подразделения и сотрудники).

## Описание

REST API для работы с:
- **Подразделениями (Departments)** — иерархическая структура с поддержкой вложенности
- **Сотрудниками (Employees)** — привязка сотрудников к подразделениям

### Возможности
- Создание/чтение/обновление/удаление подразделений
- Иерархия подразделений (дерево вложенности)
- Добавление сотрудников в подразделения
- Каскадное удаление или переназначение при удалении подразделения
- Валидация данных (уникальность имён, защита от циклов)

## Быстрый старт

### Требования
- Docker и Docker Compose
- Go 1.21+ (для локальной разработки)

### Запуск через Docker Compose

```bash
docker-compose up --build
```

Сервер будет доступен на `http://localhost:8080`

### Локальный запуск

1. Запустите PostgreSQL:
```bash
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=postgres \
  -p 5432:5432 \
  postgres:15-alpine
```

2. Настройте переменные окружения:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=postgres
export SERVER_PORT=8080
```

3. Запустите приложение:
```bash
go run cmd/api/main.go
```

## API Endpoints

### Подразделения

#### Создать подразделение
```bash
POST /departments/
Content-Type: application/json

{
  "name": "Engineering",
  "parent_id": null  // опционально, ID родительского подразделения
}
```

**Ответ:** `201 Created` с объектом подразделения

#### Получить подразделение
```bash
GET /departments/{id}?depth=1&include_employees=true
```

Параметры:
- `depth` (int, 1-5) — глубина вложенности дочерних подразделений (по умолчанию 1)
- `include_employees` (bool) — включать ли сотрудников (по умолчанию true)

**Ответ:** `200 OK`
```json
{
  "id": 1,
  "name": "Engineering",
  "parent_id": null,
  "created_at": "2024-01-01T00:00:00Z",
  "employees": [...],
  "children": [...]
}
```

#### Обновить подразделение
```bash
PATCH /departments/{id}
Content-Type: application/json

{
  "name": "Engineering Updated",
  "parent_id": 2
}
```

**Ответ:** `200 OK` с обновлённым объектом

#### Удалить подразделение
```bash
DELETE /departments/{id}?mode=cascade
```

Параметры:
- `mode` — `cascade` (удалить всё) или `reassign` (переназначить сотрудников)
- `reassign_to_department_id` — обязательно при `mode=reassign`

**Ответ:** `204 No Content`

---

### Сотрудники

#### Создать сотрудника
```bash
POST /departments/{id}/employees/
Content-Type: application/json

{
  "full_name": "John Doe",
  "position": "Senior Developer",
  "hired_at": "2024-01-15"  // опционально, формат YYYY-MM-DD
}
```

**Ответ:** `201 Created` с объектом сотрудника

## Структура БД

### departments
| Поле | Тип | Описание |
|------|-----|----------|
| id | SERIAL | Первичный ключ |
| name | VARCHAR(200) | Название (не пустое) |
| parent_id | INT NULL | Ссылка на родительское подразделение |
| created_at | TIMESTAMP | Дата создания |

### employees
| Поле | Тип | Описание |
|------|-----|----------|
| id | SERIAL | Первичный ключ |
| department_id | INT | Ссылка на подразделение |
| full_name | VARCHAR(200) | ФИО (не пустое) |
| position | VARCHAR(200) | Должность (не пустая) |
| hired_at | DATE NULL | Дата приёма на работу |
| created_at | TIMESTAMP | Дата создания |

## Бизнес-правила

1. **Название подразделения:**
   - Не пустое, 1-200 символов
   - Пробелы по краям обрезаются
   - Уникально в пределах одного родителя

2. **Данные сотрудника:**
   - `full_name` и `position` не пустые, 1-200 символов
   - `hired_at` опционально, формат YYYY-MM-DD

3. **Иерархия:**
   - Нельзя сделать подразделение родителем самого себя
   - Нельзя создать цикл в дереве (возвращает `409 Conflict`)

4. **Удаление:**
   - `cascade` — удаляет подразделение, сотрудников и все дочерние подразделения
   - `reassign` — удаляет подразделение, сотрудники переводятся в указанное подразделение

## Тесты

### Unit тесты

```bash
go test ./... -v
```

### Интеграционные тесты

Для запуска интеграционных тестов требуется Docker (testcontainers):

```bash
# Запустить все интеграционные тесты
go test ./... -tags=integration -v

# Запустить тесты конкретного пакета
go test ./internal/repository/... -tags=integration -v
go test ./internal/service/... -tags=integration -v

# Запустить с покрытием
go test ./... -tags=integration -cover
```

Интеграционные тесты автоматически запускают PostgreSQL в Docker контейнере через testcontainers-go.

## Архитектура проекта

```
.
├── cmd/
│   └── api/
│       └── main.go          # Точка входа, инициализация
├── internal/
│   ├── config/
│   │   └── config.go        # Конфигурация приложения
│   ├── handler/
│   │   ├── handler.go       # HTTP обработчики
│   │   └── handler_test.go  # Тесты обработчиков
│   ├── model/
│   │   └── model.go         # Модели данных и DTO
│   ├── repository/
│   │   └── repository.go    # Работа с БД (GORM)
│   └── service/
│       └── service.go       # Бизнес-логика
├── migrations/
│   └── 20260220143921_initial_schema.sql  # Миграции БД
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## Технологии

- **Go 1.24+** — язык программирования
- **net/http** — стандартная библиотека для HTTP сервера
- **GORM** — ORM для работы с PostgreSQL
- **Goose** — миграции базы данных
- **PostgreSQL** — основная база данных
- **slog** — структурированное логирование (стандартная библиотека)
- **Docker & Docker Compose** — контейнеризация
- **testcontainers-go** — интеграционные тесты с PostgreSQL в Docker

## Примеры использования

### Создание структуры подразделений

```bash
# Создать корневое подразделение
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Company"}'

# Создать дочернее подразделение
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Engineering", "parent_id": 1}'

# Добавить сотрудника
curl -X POST http://localhost:8080/departments/2/employees/ \
  -H "Content-Type: application/json" \
  -d '{"full_name": "John Doe", "position": "Developer"}'

# Получить дерево подразделений
curl http://localhost:8080/departments/1?depth=3
```

## Конфигурация

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_HOST` | Хост PostgreSQL | localhost |
| `DB_PORT` | Порт PostgreSQL | 5432 |
| `DB_USER` | Пользователь БД | postgres |
| `DB_PASSWORD` | Пароль БД | postgres |
| `DB_NAME` | Имя базы данных | postgres |
| `SERVER_PORT` | Порт HTTP сервера | 8080 |

## License

MIT