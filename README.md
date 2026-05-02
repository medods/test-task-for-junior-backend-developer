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

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

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
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

## Настройки периодичности задач

В модели задачи добавлено поле `recurrence`.

Поддерживаемые типы:

- `none` — без периодичности (значение по умолчанию)
- `daily` — каждый `n`-й день, поле `every_n_days` (диапазон `1..365`)
- `monthly` — каждый месяц в конкретное число, поле `day_of_month` (диапазон `1..30`)
- `specific_dates` — задачи на конкретные даты, поле `dates` в формате `YYYY-MM-DD`
- `month_day_parity` — только на четные или нечетные дни месяца, поле `parity` (`even`/`odd`)

Пример payload с периодичностью:

```json
{
  "title": "Пациентский обзвон",
  "description": "Проверить состояние после выписки",
  "status": "new",
  "recurrence": {
    "type": "daily",
    "every_n_days": 2
  }
}
```

## Принятые предположения по заданию

- Периодичность хранится как часть сущности задачи в поле `recurrence`.
- Если `recurrence` не передан, используется `{"type":"none"}`.
- Для `none` дополнительные поля периодичности должны быть пустыми.
- Для `specific_dates` даты валидируются, дубликаты удаляются, итоговый список хранится в отсортированном виде.
- Для `monthly` допускаются только числа `1..30`, как указано в задании.
- Периодичность влияет на правила задачи и хранится в API/БД, дальнейшая автоматическая генерация отдельных экземпляров задач (по расписанию фоновым процессом) в рамках этого задания не добавлялась.
