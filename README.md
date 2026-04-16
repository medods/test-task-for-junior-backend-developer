# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.25+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Compose поднимает `postgres`, затем выполняет миграции (`migrate up`) и сидирование (`seed`).

Если `postgres` уже запускался ранее, пересоздай volume, чтобы получить чистую БД:

```bash
docker compose down -v
docker compose up --build
```

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks?date=YYYY-MM-DD`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `PATCH /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Миграции и сидирование

Есть отдельные утилиты для миграций и сидирования базы в `cmd/`.

- Применить миграции (локально):

```bash
DATABASE_DSN="postgres://postgres:postgres@localhost:5430/taskservice?sslmode=disable" go run ./cmd/migrate up
```

- Применить миграции лестницей (staircase):

```bash
DATABASE_DSN="postgres://postgres:postgres@localhost:5430/taskservice?sslmode=disable" go run ./cmd/migrate staircase
```

- Сидирование базы (если таблица `tasks` пуста):

```bash
DATABASE_DSN="postgres://postgres:postgres@localhost:5430/taskservice?sslmode=disable" go run ./cmd/seed
```

Также эти команды доступны в контейнерах через `docker compose run`, например:

```bash
docker compose run --rm migrate /app/migrate up
docker compose run --rm seed /app/seed
```

Процесс миграций в CI/контейнерной среде реализован как "staircase" и оркестрируется с помощью `seqwall` — в проекте есть собранный образ мигратора (`docker/migrator/Dockerfile`) и скрипты, которые последовательно выполняют проверочные шаги (actualise, down-up-down, reapply) перед финальным применением миграций к базе.

## Интеграционные тесты

Интеграционные тесты против Postgres запускаются только если задана переменная окружения `TEST_DATABASE_DSN`.
Пример запуска (после поднятия `postgres` через `docker compose up -d postgres`):

```bash
TEST_DATABASE_DSN="postgres://postgres:postgres@localhost:5430/taskservice?sslmode=disable" go test ./... -v
```

После тестов можно очистить том:

```bash
docker compose down -v
```

## Разработка

- Быстрый локальный запуск (скрипт и Makefile):

```bash
./run.sh     # запустит `docker compose up --build`
make up      # тоже самое через Makefile
make start   # поднять в background и показать логи
make down    # остановить и удалить контейнеры
```

- Тесты:

```bash
make test             # unit tests: go test ./... -cover
make test-integration # интеграционные тесты (поднимают postgres + migrator)
```


## Фичa: Периодичность задач (recurrence)

Добавлены параметры периодичности задач и соответствующие проверки:

- `None` — одноразовая задача
- `DailyEveryN` — каждый N-й день (минимум 1)
- `MonthlyDay` — конкретный день месяца (1..30)
- `SpecificDates` — набор конкретных дат
- `MonthParity` — чётные / нечётные дни

API заметки:

- `GET /api/v1/tasks?date=YYYY-MM-DD` — получить задачи, активные на указанную дату
- `PATCH /api/v1/tasks/{id}` — частичное обновление (поддерживается явное `null` для `recurrence_start_date`)

Валидация входящих запросов реализована в HTTP-слое (validator) и дополнительно проверяется в usecase при логической валидации recurrence-полей.

## Прочее

- Включён простой rate-limiter на HTTP-маршрутах (возвращает `429` при превышении).
- Для разработки: `go test ./...`, `go vet ./...`, `go build ./...` — все пройдут при корректной среде.
