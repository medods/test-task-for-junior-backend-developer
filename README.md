# Task Service

Сервис для управления разовыми и периодическими задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска:

- API будет доступен по адресу `http://localhost:8080`
- `postgres` инициализирует схему из bootstrap SQL
- `scheduler` выполнит первичную материализацию периодических задач и затем будет запускать её раз в день

## Сервисы Docker Compose

- `postgres` хранит данные и применяет bootstrap SQL при первом запуске на пустом volume.
- `app` запускает HTTP API и Swagger UI.
- `scheduler` запускает одноразовый `task-materializer` при старте контейнера и затем повторяет materialization по cron.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что bootstrap SQL из `migrations/*.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

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

Маршруты задач:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

Маршруты расписаний:

- `POST /api/v1/task-schedules`
- `GET /api/v1/task-schedules`
- `GET /api/v1/task-schedules/{id}`
- `PUT /api/v1/task-schedules/{id}`
- `DELETE /api/v1/task-schedules/{id}`

## Что сделано и почему

- Периодичность вынесена в отдельную сущность `TaskSchedule`, а `Task` остался экземпляром выполнения. Это разделяет правило генерации и факт выполнения задачи.
- Генерация задач вынесена из `GET /tasks`. Чтение остается чтением, а materialization выполняется при создании/изменении расписания и отдельным `scheduler`-сервисом.
- В БД используется `rule_type + rule_payload JSONB`. В Go домен остается строго типизированным, а в схему БД не нужно вносить изменения при добавлении будущего `weekly`.
- `task-materializer` сделан одноразовой командой. Расписание запуска задает инфраструктура через `scheduler` в `docker compose`, а не вечный цикл в бизнес-логике.

## Разумные предположения

- `weekly` в v1 не реализован, но модель подготовлена к его добавлению.
- Окно материализации по умолчанию: `today..today+30d`.
- Все сгенерированные задачи создаются в статусе `new`.
- При изменении расписания пересобираются только будущие задачи этого расписания со статусом `new`.
- `DELETE /api/v1/task-schedules/{id}` означает деактивацию расписания, а не физическое удаление истории.
- `monthly_day` ограничен значениями от `1` до `30`.
- Bootstrap-схема гарантирована только для новой БД на пустом volume.
