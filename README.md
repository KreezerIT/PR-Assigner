# PR-Assigner Service

API для автоматического назначения ревьюверов на Pull Request'ы с управлением командами и участниками

## Оглавление

- [Принципы проектирования](#принципы-проектирования)
- [Технологический стек](#технологический-стек)
- [Быстрый запуск](#быстрый-запуск)
  - [Склонировать репозиторий](#склонировать-репозиторий)
  - [Запуск через Docker Compose](#запуск-через-docker-compose)
  - [Тестирование](#тестирование)
  - [Остановка сервисов](#остановка-сервисов)
  - [Makefile команды](#makefile-команды)
  - [Линтер](#линтер)
  - [Проверка работоспособности (API Endpoints) через curl](#проверка-работоспособности-api-endpoints-через-curl)

### Принципы проектирования

- **High Cohesion, Low Coupling** - каждый модуль отвечает за одну доменную сущность
- **Dependency Inversion** - handlers зависят от интерфейсов, а не конкретных реализаций
- **Single Responsibility** - четкое разделение ответственности между слоями
- **DTO Pattern** - раздельные DTO для HTTP и Domain слоев
- **Repository Pattern** - абстракция работы с данными
- **Validator Pattern (Composite)** - композитная валидация без множественных if

## Технологический стек

- **Язык**: Go 1.23
- **Web**: Echo v4
- **БД**: PostgreSQL 15, Squirrel
- **Миграции**: golang-migrate/migrate
- **Логирование**: Zap (uber-go/zap)
- **Метрики**: Prometheus
- **Тестирование**: testify, httptest
- **Контейнеризация**: Docker, Docker Compose
- **Линтер**: golangci-lint

## Быстрый запуск

### Склонировать репозиторий и .env файл

```bash
git clone https://github.com/yourusername/pr-reviewer-service.git
cd pr-reviewer-service
```
```bash
cp .env.example .env
```
### Запуск через Docker Compose
```bash
# Запуск всех сервисов (PostgreSQL + миграции + API)
docker-compose up --build

# Или через Makefile
make docker-up
```

### Тестирование
```bash
# Все тесты (unit + integration)
make test
```
`Unit-тесты`
```bash
# Запуск всех unit-тестов
make test-unit

# Или напрямую
go test ./...
```
```bash
# Или напрямую
go test -v -race -coverprofile=coverage.out ./internal/...
```
```bash
# Просмотр покрытия бизнес-логики
make coverage

# Или напрямую (директории с бизнес-логикой)
go test -coverprofile="coverage.out" ./internal/app/handler/pullrequest ./internal/app/handler/statistics ./internal/app/handler/team ./internal/app/handler/user ./internal/app/validator ./internal/domain/...
```

`Интеграционные тесты`
```bash
# Запуск интеграционных тестов (автоматически поднимает тестовую БД)
make test-integration

# Или напрямую
go test ./tests/integration
```

### Остановка сервисов
```bash
docker-compose down -v

# Или через Makefile
make docker-down
```
*Сервис будет доступен на http://localhost:8080*

### Makefile команды

- `make build` - собрать бинарный файл приложения
- `make run` - запустить приложение локально
- `make docker-test` - запустить все тесты внутри Docker
- `make migrate-create NAME=...` - создать новый файл миграции
- `make migrate-up` - применить все новые миграции
- `make migrate-down` - откатить последнюю примененную миграцию
- `make lint` - запустить статический анализ кода
- `make deps` - загрузить и упорядочить зависимости проекта

### Линтер
Конфиг находится в `.golangci.yml`
```bash
make lint

# Или напрямую
golangci-lint run --timeout 5m
```
### Проверка работоспособности (API Endpoints) через curl

#### /health (Health check)
```bash
curl http://localhost:8080/health
```
```json
{"status":"ok"}
```

#### /metrics (Prometheus метрики)
```bash
curl http://localhost:8080/metrics
```

Дальнейший формат:
- "Endpoint"
- Пример кода для вызова (*Linux/macOS*)
- Пример кода для вызова (*Windows CMD*)
- Код успешного ответа
- Пример ожидаемого ответа


`TEAM`
#### POST /team/add (Создать команду)
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }'
```
```bash
curl.exe -X POST http://localhost:8080/team/add -H "Content-Type: application/json" -d "{\""team_name\"":\""backend\"",\""members\"":[{\""user_id\"":\""u1\"",\""username\"":\""Alice\"",\""is_active\"":true},{\""user_id\"":\""u2\"",\""username\"":\""Bob\"",\""is_active\"":true},{\""user_id\"":\""u3\"",\""username\"":\""Charlie\"",\""is_active\"":true}]}"
```
201 Created
```json
{
  "team": {
    "team_name": "backend",
    "members": [
      {
        "user_id": "u1",
        "username": "Alice",
        "is_active": true
      },
      {
        "user_id": "u2",
        "username": "Bob",
        "is_active": true
      },
      {
        "user_id": "u3",
        "username": "Charlie",
        "is_active": true
      }
    ]
  }
}
```
#### GET /team/get?team_name=backend (Получить команду)
```bash
curl "http://localhost:8080/team/get?team_name=backend"
```
```bash
curl.exe "http://localhost:8080/team/get?team_name=backend"
```
```json
{
  "team_name": "backend",
  "members": [
    {
      "user_id": "u1",
      "username": "Alice",
      "is_active": true
    },
    {
      "user_id": "u2",
      "username": "Bob",
      "is_active": true
    }
  ]
}
```

`USER`
#### POST /users/setIsActive (Изменить статус активности пользователя)
```bash
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u2",
    "is_active": false
  }'
```
```bash
curl.exe -X POST http://localhost:8080/users/setIsActive -H "Content-Type: application/json" -d "{\""user_id\"":\""u2\"",\""is_active\"":false}"
```
200 OK
```json
{
  "user": {
    "user_id": "u2",
    "username": "Bob",
    "team_name": "backend",
    "is_active": false
  }
}
```

`PR`
#### POST /pullRequest/create (Создать Pull Request)
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "pull_request_name": "Add amazing feature",
    "author_id": "u1"
  }'
```
```bash
curl.exe -X POST http://localhost:8080/pullRequest/create -H "Content-Type: application/json" -d "{\""pull_request_id\"":\""pr-1\"",\""pull_request_name\"":\""Add amazing feature\"",\""author_id\"":\""u1\""}"
```
201 Created
```json
{
  "pr": {
    "pull_request_id": "pr-1",
    "pull_request_name": "Add amazing feature",
    "author_id": "u1",
    "status": "OPEN",
    "assigned_reviewers": ["u2", "u3"],
    "createdAt": "2025-01-20T12:00:00Z"
  }
}
```

#### POST /pullRequest/merge (Слить Pull Request)
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1"
  }'
```
```bash
curl.exe -X POST http://localhost:8080/pullRequest/merge -H "Content-Type: application/json" -d "{\""pull_request_id\"":\""pr-1\""}"
```
200 OK
```json
{
  "pr": {
    "pull_request_id": "pr-1",
    "pull_request_name": "Add amazing feature",
    "author_id": "u1",
    "status": "MERGED",
    "assigned_reviewers": ["u2", "u3"],
    "createdAt": "2025-01-20T12:00:00Z",
    "mergedAt": "2025-01-20T14:30:00Z"
  }
}
```

#### POST /pullRequest/reassign (Переназначить ревьювера)
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "old_user_id": "u2"
  }'
```
```bash
curl.exe -X POST http://localhost:8080/pullRequest/reassign -H "Content-Type: application/json" -d "{\""pull_request_id\"":\""pr-2\"",\""old_user_id\"":\""u3\""}"
```
200 OK
```json
{
  "pr": {
    "pull_request_id": "pr-1",
    "pull_request_name": "Add amazing feature",
    "author_id": "u1",
    "status": "OPEN",
    "assigned_reviewers": ["u3", "u4"]
  },
  "replaced_by": "u4"
}
```

#### GET /users/getReview?user_id=u2 (Получить PR пользователя)
```bash
curl "http://localhost:8080/users/getReview?user_id=u2"
```
```bash
curl.exe "http://localhost:8080/users/getReview?user_id=u2"
```
```json
{
  "user_id": "u2",
  "pull_requests": [
    {
      "pull_request_id": "pr-1",
      "pull_request_name": "Add amazing feature",
      "author_id": "u1",
      "status": "OPEN"
    },
    {
      "pull_request_id": "pr-3",
      "pull_request_name": "Fix critical bug",
      "author_id": "u5",
      "status": "MERGED"
    }
  ]
}
```

`STATISTICS`
#### GET /statistics/global (Глобальная статистика)
```bash
curl.exe http://localhost:8080/statistics/global
```
200 OK
```json
{
  "total_users": 50,
  "active_users": 45,
  "total_teams": 5,
  "total_prs": 100,
  "open_prs": 20,
  "merged_prs": 80,
  "total_assignments": 150
}
```

#### GET /statistics/user/:user_id (Статистика по пользователю)
```bash
curl http://localhost:8080/statistics/user/u2
```
200 OK
```json
{
  "user_id": "u2",
  "username": "Bob",
  "team_name": "backend",
  "total_assignments": 10,
  "active_reviews": 3,
  "completed_reviews": 7
}
```

#### GET /statistics/team/:team_name (Статистика по команде)
```bash
curl http://localhost:8080/statistics/team/backend
```
200 OK
```json
{
  "team_name": "backend",
  "total_members": 10,
  "active_members": 8,
  "total_prs": 50,
  "open_prs": 10,
  "merged_prs": 40
}
```

#### GET /statistics/top-reviewers?limit=10 (Топ ревьюверов)
```bash
curl.exe "http://localhost:8080/statistics/top-reviewers?limit=10"
```
```json
{
  "top_reviewers": [
    {
      "user_id": "u2",
      "username": "Bob",
      "team_name": "backend",
      "review_count": 25
    },
    {
      "user_id": "u5",
      "username": "Eve",
      "team_name": "frontend",
      "review_count": 20
    },
    {
      "user_id": "u3",
      "username": "Charlie",
      "team_name": "backend",
      "review_count": 18
    }
  ]
}
```

