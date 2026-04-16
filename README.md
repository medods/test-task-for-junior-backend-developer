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

Если `postgres` уже запускался ранее, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файлы из `migrations/` монтируются в `docker-entrypoint-initdb.d` и применяются только при инициализации пустого data volume.

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
- `GET /api/v1/tasks?recurrence_group_id=<uuid>`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`
- `POST /api/v1/tasks/recurring`
- `DELETE /api/v1/tasks/recurring/{group_id}`

## Принятые решения и предположения

При создании серии система разворачивает правило повторяемости в набор дат и создаёт одну задачу на каждую дату в одной транзакции. Все задачи серии связаны общим `recurrence_group_id`. Каждый экземпляр — независимая задача со своим статусом.

Поддерживаемые типы повторяемости:

| Тип | Обязательные поля |
|---|---|
| `daily` | `every_n_days` (≥1), `start_date`, `end_date` |
| `monthly` | `days_of_month` (1–30 каждое), `start_date`, `end_date` |
| `specific_dates` | `dates` (массив строк `YYYY-MM-DD`) |
| `even_odd` | `parity` (`even`/`odd`), `start_date`, `end_date` |

Ограничения и граничные случаи:

- Максимум 366 экземпляров на один запрос.
- Числа месяца, которых нет в конкретном месяце (например, 30 февраля), пропускаются.
- Дубликаты в `specific_dates` дедуплицируются.
- `days_of_month` ограничен значениями 1–30 согласно условию задачи.

Прочее:

- `scheduled_at` добавлен ко всем задачам — одиночные задачи тоже могут иметь срок выполнения.
- Статус `canceled` добавлен — присутствует в UI-макете, нужен для отмены отдельных экземпляров серии.
- Изменение правила повторяемости не поддерживается — для этого нужно удалить серию и создать новую. Так сохраняется история уже выполненных задач.
- `DELETE /tasks/recurring/{group_id}` физически удаляет задачи. Для «мягкой» отмены достаточно обновить статус каждого экземпляра через `PUT /tasks/{id}`.
