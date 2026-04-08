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

Периодичность задач
Сервис поддерживает следующие типы периодичности для задач:

- `Ежедневные (daily) - задача повторяется каждый n-й день`
- `Ежемесячные (monthly_day) - задача выполняется каждый месяц в определенное число (1-30)`
- `Конкретные даты (specific) - задача выполняется только в указанные даты`
- `Четные/нечетные дни (even_odd) - задача выполняется только в четные или нечетные дни месяца`

Примеры создания задач с периодичностью
1. Обычная задача (без периодичности)

```text
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Купить продукты",
    "description": "Молоко, хлеб, яйца",
    "status": "new"
  }'
```

2. Ежедневная задача (каждый 3-й день)

```text
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Полить цветы",
    "description": "Поливать комнатные растения",
    "status": "new",
    "recurrence_type": "daily",
    "recurrence_every_n": 3
  }'
```

3. Ежемесячная задача (каждое 15-е число)

```text
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Сдать отчет",
    "description": "Ежемесячный финансовый отчет",
    "status": "new",
    "recurrence_type": "monthly_day",
    "monthly_day": 15
  }'
```

4. Задача на конкретные даты

```text
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Встреча с клиентом",
    "description": "Обсуждение нового проекта",
    "status": "new",
    "recurrence_type": "specific",
    "specific_dates": ["2026-04-15T10:00:00Z", "2026-04-20T14:00:00Z", "2026-04-25T11:00:00Z"]
  }'
```

5. Задача на четные дни месяца

```text
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Резервное копирование",
    "description": "Делать бэкап в четные дни",
    "status": "new",
    "recurrence_type": "even_odd",
    "even_odd": "even"
  }'
```

6. Задача на нечетные дни месяца

```text
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Генеральная уборка",
    "description": "Проводить уборку в нечетные дни",
    "status": "new",
    "recurrence_type": "even_odd",
    "even_odd": "odd"
  }'
```

Пример обновления задачи с периодичностью
```text
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Обновленный заголовок",
    "recurrence_type": "daily",
    "recurrence_every_n": 5
  }'
```