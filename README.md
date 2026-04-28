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

## Что реализовано

Поддерживаются три режима задачи:

- `one_time` - обычная разовая задача.
- `recurring_template` - шаблон периодической задачи.
- `recurring_instance` - конкретный экземпляр периодической задачи на дату на основе шаблона.

Поддерживаются пять типов периодичности:

- `daily` - каждый `n`-й день.
- `monthly` - в определённый день месяца от `1` до `30`.
- `specific_dates` - только на конкретные даты.
- `odd_days` - только по нечётным дням месяца.
- `even_days` - только по чётным дням месяца.

## Как это реализовано

Периодические задачи реализованы через модель `template + instances`:

- шаблон хранит правило повторения;
- реальные задачи на конкретные даты создаются как `recurring_instance`;
- у каждого инстанса свой собственный `status`;
- завершение одного инстанса не ломает будущие повторения.

При создании `recurring_template` сервис сразу генерирует инстансы на `30` дней вперёд.  
При `GET /tasks` сервис сначала синхронизирует шаблоны и достраивает недостающие инстансы на тот же горизонт.

## Особенности поведения

- `GET /tasks` возвращает все записи: и `one_time`, и `recurring_template`, и `recurring_instance`.
- `mode=recurring_instance` нельзя создать напрямую через `POST /tasks`.
- `mode` нельзя менять через `PUT /tasks/{id}`.
- Для `recurring_instance` нельзя менять настройки повторов.
- Новые сгенерированные инстансы всегда создаются со статусом `new`.
- Все повторные даты передаются в формате `YYYY-MM-DD`.
- Внутри приложения повторные даты нормализуются в `UTC midnight` из-за временных зон.

## Крайние случи

### 1. Короткие месяцы и `day_of_month = 30`

Для `monthly` число `30` допустимо всегда.  
Если в месяце такого дня нет (привет, февраль), используется последний день месяца.

Примеры:

- февраль 2026 -> `YYYY-02-28`
- февраль високосного года -> `YYYY-02-29`

### 2. `specific_dates` с датами в прошлом

Такие даты разрешены и сохраняются в шаблоне.  
Но инстансы создаются только для дат, попадающих в текущее окно генерации.

### 3. Дубликаты в `specific_dates`

Дубли запрещены.  
Если в массиве `dates` передать одну и ту же дату дважды, запрос считается невалидным.

### 4. Обновление шаблона

При обновлении шаблона:

- сам шаблон обновляется;
- будущие инстансы со статусом `new` пересобираются заново;
- инстансы со статусами `in_progress` и `done` не удаляются.

Это позволяет менять правило повторения без потери уже начатой или завершённой задачи.

### 5. Удаление шаблона

При удалении шаблона:

- сначала удаляются дочерние инстансы со статусом `new`;
- затем сервис пытается удалить сам шаблон;
- если остаются связанные инстансы не в статусе `new`, удаление завершается ошибкой `409 Conflict`.

Защита истории задач.

### 6. Нечётные и чётные дни

Для `odd_days` и `even_days` обязательно нужен `start_date`.  
Генерация начинается не раньше этой даты.

## Валидация повторов

Правила валидации:

- `daily`:
  - обязателен `every_n_days > 0`
  - обязателен `start_date`
  - запрещены `day_of_month` и `dates`
- `monthly`:
  - обязателен `day_of_month` в диапазоне `1..30`
  - обязателен `start_date`
  - запрещены `every_n_days` и `dates`
- `specific_dates`:
  - обязателен непустой массив `dates`
  - `dates` должны быть уникальными
  - `start_date` запрещён
  - `every_n_days` и `day_of_month` запрещены
- `odd_days`:
  - обязателен `start_date`
  - запрещены `every_n_days`, `day_of_month`, `dates`
- `even_days`:
  - обязателен `start_date`
  - запрещены `every_n_days`, `day_of_month`, `dates`

## Особенности реализации

- Шаблоны и инстансы задач хранятся в одной таблице `tasks`.
- Связь инстанса с шаблоном через `source_task_id`.
- Дата конкретного инстанса в `scheduled_for`.
- Для защиты от повторов действует уникальность пары:
  - `source_task_id + scheduled_for` для `recurring_instance`.
- Для `recurrence_dates` в БД есть проверка:
  - массив одномерный
  - не пустой
  - без `NULL`
  - без повторов

Репозиторий разрезан:

- `task_repository.go` - базовый CRUD.
- `task_scan.go` - чтение и сборка доменной модели.
- `task_params.go` - подготовка параметров записи.
- `task_recurring_repository.go` - операции над шаблонами.
- `task_instance_repository.go` - операции над инстансами.

## Примеры

### Обычная задача

```json
{
  "title": "Prepare food",
  "description": "Collect food and do magic",
  "status": "new"
}
```

### Ежедневный шаблон повторов

```json
{
  "title": "Call patients",
  "description": "Daily follow-up calls",
  "status": "new",
  "mode": "recurring_template",
  "recurrence": {
    "type": "daily",
    "every_n_days": 1,
    "start_date": "2026-04-28"
  }
}
```

### Месячные повторы

```json
{
  "title": "Prepare monthly report",
  "description": "Collect data",
  "status": "new",
  "mode": "recurring_template",
  "recurrence": {
    "type": "monthly",
    "day_of_month": 30,
    "start_date": "2026-04-28"
  }
}
```

### Повторы по датам

```json
{
  "title": "Inventory check",
  "description": "Check drug store",
  "status": "new",
  "mode": "recurring_template",
  "recurrence": {
    "type": "specific_dates",
    "dates": ["2026-05-01", "2026-05-10", "2026-05-25"]
  }
}
```

## Тесты

В проекте есть unit-тесты для usecase-слоя.

Покрыты:

- date helper-ы:
  - `dateOnly`
  - `clampedDateInMonth`
  - `daysInMonth`
- recurrence-логика:
  - normalizing и сортировка `specific_dates`
  - запрет дублей в `specific_dates`
  - обязательность `start_date`
  - генерация дат для `daily`
  - генерация дат для `monthly`
  - генерация дат для `specific_dates`
  - генерация дат для `odd_days` и `even_days`
- service orchestration:
  - create для `one_time`
  - create для `recurring_template`
  - запрет прямого create для `recurring_instance`
  - update для `recurring_template`
  - запрет recurrence-update для `recurring_instance`
  - запрет смены `mode`
  - delete для `recurring_template`
  - sync шаблонов в `List`

## Как запускать тесты

```cmd
go test ./...
```
