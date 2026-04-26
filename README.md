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

## Периодичность задач

### Маршруты

- `GET /api/v1/tasks/{id}/recurrence` — получить настройки периодичности
- `PUT /api/v1/tasks/{id}/recurrence` — установить или заменить периодичность
- `DELETE /api/v1/tasks/{id}/recurrence` — удалить периодичность

### Типы периодичности

| `type` | Доп. поле | Описание |
|---|---|---|
| `daily` | `interval: N` | Каждые N дней |
| `monthly` | `day_of_month: N` | N-го числа каждого месяца (1-30) |
| `weekly` | `week_days: [1..7]` | По указанным дням недели (1=пн, 7=вс) |
| `specific_dates` | `dates: ["YYYY-MM-DD"]` | Только в указанные даты |

Примеры запросов:

```json
// Каждые 2 дня
PUT /api/v1/tasks/1/recurrence
{ "type": "daily", "interval": 2 }

// 15-го числа каждого месяца
PUT /api/v1/tasks/1/recurrence
{ "type": "monthly", "day_of_month": 15 }

// По понедельникам, средам и пятницам
PUT /api/v1/tasks/1/recurrence
{ "type": "weekly", "week_days": [1, 3, 5] }

// Только в конкретные даты
PUT /api/v1/tasks/1/recurrence
{ "type": "specific_dates", "dates": ["2026-05-01", "2026-06-12"] }
```

### Поведение

- Периодичность можно установить только на уже существующую задачу.
- В ответе `GET /api/v1/tasks/{id}` появляется поле `recurrence_id` (null если периодичности нет).
- Когда задача с периодичностью переведена в статус `done`, она **остаётся в `done`** до наступления следующей даты по расписанию.
- Планировщик запускается каждую ночь в полночь: находит задачи со статусом `done`, у которых наступила дата `next_occurrence_at`, и сбрасывает их обратно в `new`.
- Для типа `specific_dates`: когда все даты из списка остались в прошлом, задача больше не сбрасывается автоматически.

### Архитектурные решения

- **Отдельная таблица `task_recurrences`** со столбцами `type TEXT` и `config JSONB`. `type` вынесен отдельно для читаемости и DB-level CHECK-валидации; `config` хранит только специфичные для типа поля — нет лишних NULL-колонок.
- **`tasks.recurrence_id`** — nullable FK на `task_recurrences`. Задача владеет recurrence, а не наоборот: нет цикличных зависимостей, `recurrence_id IS NULL` заменяет отдельный boolean-флаг.
- **Интерфейс `Recurrence`** с отдельной структурой под каждый тип (`DailyRecurrence`, `MonthlyRecurrence`, `WeeklyRecurrence`, `SpecificDatesRecurrence`). Каждый тип реализует `Valid()` и `NextOccurrence()` — добавление нового типа не затрагивает существующий код.
- **Планировщик** реализован как горутина внутри того же процесса. Метод `Tick` экспортирован для возможности тестирования без ожидания полночи.
