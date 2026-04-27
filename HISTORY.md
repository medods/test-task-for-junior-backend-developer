# для чего нам нужны эти импорты

```
taskdomain "example.com/taskservice/internal/domain/task"
taskusecase "example.com/taskservice/internal/usecase/task"
```

Эти импорты подключают два внутренних пакета и дают им удобные псевдонимы:

* `taskdomain` — пакет домена (entities/ошибки). В нём, например, определено `ErrNotFound` и типы сущности задачи.
* `taskusecase` — пакет бизнес-логики (use-cases). В нём объявлен интерфейс `Usecase`, структуры входных данных (`CreateInput`, `UpdateInput`) и ошибки уровня usecase (`ErrInvalidInput`).

Зачем нужны:

* В `TaskHandler` поле `usecase` имеет тип `taskusecase.Usecase`.
* В методах создаются/обновляются данные через `taskusecase.CreateInput` / `UpdateInput`.
* В `writeUsecaseError` сравниваются ошибки `taskdomain.ErrNotFound` и `taskusecase.ErrInvalidInput`.

Псевдонимы нужны, чтобы:

* избежать конфликта имён (оба пакета могут быть `task`),
* сделать код более явным (чётко видно, откуда тип/ошибка).

---

# Сгенерируй body для post запросов

Example JSON bodies for `POST /api/v1/tasks` (create task). Fields used: `title` (required), `description`, `dueDate` (RFC3339), `completed` (bool), `priority` (int), `tags` (array of strings).

### minimal

```json
{
  "title": "Buy groceries"
}
```

### full

```json
{
  "title": "Finish report",
  "description": "Complete monthly sales report and send to stakeholders.",
  "dueDate": "2026-05-10T17:00:00Z",
  "completed": false,
  "priority": 2,
  "tags": ["work", "reports"]
}
```

### with_due_date_only

```json
{
  "title": "Pay rent",
  "dueDate": "2026-05-01T12:00:00Z"
}
```

### with_tags

```json
{
  "title": "Plan team offsite",
  "description": "Choose date and venue, prepare agenda.",
  "tags": ["team", "planning", "events"]
}
```

### invalid_missing_title

```json
{
  "description": "This body is invalid because title is required."
}
```
