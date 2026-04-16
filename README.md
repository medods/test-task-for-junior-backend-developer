# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
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
