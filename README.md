# Team Task Management Service

REST API для управления задачами в командах: роли участников, история изменений задач, кеширование, rate limiting, circuit breaker, Prometheus-метрики.

Стек: Go, MySQL, Redis, Docker + Docker Compose.

## Структура репозитория

```
app/      — весь Go-код, миграции, конфиги приложения (app/.env, config.yaml)
docker/   — Dockerfile'ы, docker-compose.yml, volume для контейнеров (docker/.env)
```

## 1. Установка / первый запуск

### Требования

- Docker + Docker Compose
- (опционально) Go 1.25+ — если нужно запускать `go test`/`go build` локально, вне контейнера

### Шаги

1. Скопировать примеры конфигов и при необходимости поправить значения:

   ```bash
   cp app/.env.example app/.env
   cp docker/.env.example docker/.env
   ```

2. Поднять стек (MySQL + Redis + приложение):

   ```bash
   docker compose -f docker/docker-compose.yml up -d --build
   ```

   Миграции применяются автоматически при старте приложения (golang-migrate, см. `app/migrations`).

3. Проверить, что сервис поднялся:

   ```bash
   curl http://localhost:8080/healthz
   # {"status":"ok"}
   ```

### Режим разработки (hot-reload)

Контейнер `app` запускает [air](https://github.com/air-verse/air) — при любом изменении кода в `app/` (бинд-маунт в контейнер) или просто при `docker compose restart app` приложение пересобирается заново. Удалять и пересобирать контейнер вручную не требуется:

```bash
docker compose -f docker/docker-compose.yml restart app
```

### Конфигурация

- `app/.env` — секреты и параметры окружения приложения (БД, Redis, JWT secret). Читаются Go-процессом напрямую.
- `app/config.yaml` — несекретные дефолты: порт, пагинация, TTL кеша, rate limit, expiry JWT.
- `docker/.env` — переменные только для `docker-compose.yml` (порты, пароль root MySQL, имя БД).

Значения из `app/.env` переопределяют дефолты из `config.yaml` (если заданы).

## 2. Список API

Базовый префикс: `/api/v1`. Авторизованные эндпоинты требуют заголовок `Authorization: Bearer <token>`, полученный через `/login`.

### Аутентификация (публичные, без токена)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/v1/register` | Регистрация пользователя (`email`, `password`, `name`) |
| POST | `/api/v1/login` | Логин (`email`, `password`) → `{ "token": "<jwt>" }` |

### Команды (требуют авторизации)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/v1/teams` | Создать команду (`name`) — создатель становится `owner` |
| GET | `/api/v1/teams` | Список команд, где состоит текущий пользователь |
| POST | `/api/v1/teams/{id}/invite` | Пригласить пользователя по `email` (только `owner`/`admin`) |

### Задачи (требуют авторизации)

| Метод | Путь | Описание |
|---|---|---|
| POST | `/api/v1/tasks` | Создать задачу (`team_id`, `title`, `description`, `assignee_id?`) — только член команды |
| GET | `/api/v1/tasks?team_id=&status=&assignee_id=&page=&page_size=` | Список задач с фильтрами и пагинацией (результат кешируется в Redis на 5 минут) |
| PUT | `/api/v1/tasks/{id}` | Обновить задачу (`title?`, `description?`, `status?`, `assignee_id?`) — любой член команды, изменения логируются в историю |
| GET | `/api/v1/tasks/{id}/history` | История изменений задачи |

### Admin / сложные SQL-отчёты (требуют авторизации)

| Метод | Путь | Описание |
|---|---|---|
| GET | `/api/v1/admin/teams/stats` | По каждой команде: название, кол-во участников, кол-во задач `done` за последние 7 дней |
| GET | `/api/v1/admin/tasks/top-creators` | Топ-3 пользователя по числу созданных задач в каждой команде за текущий месяц |
| GET | `/api/v1/admin/tasks/orphaned-assignees` | Задачи, у которых assignee не является членом команды задачи (проверка целостности) |

### Технические эндпоинты

| Метод | Путь | Описание |
|---|---|---|
| GET | `/healthz` | Проверка живости сервиса |
| GET | `/metrics` | Prometheus-метрики (количество запросов, ошибок, latency) |

### Лимиты

- Rate limit: 100 запросов/минуту на пользователя (по IP — для `/register` и `/login`, по user ID — для остальных эндпоинтов). При превышении — `429 Too Many Requests`.
- Пагинация: `page` по умолчанию `1`, `page_size` по умолчанию `20`, максимум `100`.
