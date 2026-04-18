# Task Tracker API — Recurring Tasks Feature

## О проекте

Это API модуля трекера задач.  
В рамках тестового задания добавлена поддержка периодических задач.

При создании задачи можно передать настройки периодичности, после чего система создаёт обычные задачи на соответствующие даты.

## Что реализовано

Поддержаны следующие типы периодичности:

- `daily` — каждые `n` дней
- `monthly` — каждый месяц в указанное число (от 1 до 30)
- `specific_dates` — задачи на конкретные даты
- `odd_days` — задачи на нечётные дни месяца
- `even_days` — задачи на чётные дни месяца

Также сохранён базовый CRUD для обычных задач.

## Принятое решение

Периодичность реализована как правило генерации обычных задач:

- `Task` остаётся отдельной конкретной задачей
- `Recurrence` выделена как отдельная сущность в доменной модели
- генерация задач выполняется сразу при создании запроса
- каждая сгенерированная задача получает `scheduled_for`

## Допущения

- Генерация периодических задач выполняется сразу, без фоновых джоб и scheduler.
- Endpoint создания возвращает одну из созданных задач, а не весь список.
- Для `monthly` используется число месяца от 1 до 30, в соответствии с условием задания.
- Для `odd_days` и `even_days` генерация идёт по диапазону от `start_date` до `end_date`.

## Запуск проекта

1. Создать базу данных PostgreSQL.
2. По умолчанию используется строка подключения:
postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable3. Запустить приложение:
3. Запуск проекта: go run ./cmd/api

## Примеры запросов

## Примеры запросов

### Обычная задача

```json
{
  "title": "Prepare release",
  "description": "Collect release notes and check migrations",
  "status": "new",
  "scheduled_for": "2026-04-20"
}
```

specific_dates

```json
{
  "title": "Обзвон пациентов",
  "description": "specific dates test",
  "status": "new",
  "recurrence": {
    "type": "specific_dates",
    "start_date": "2026-04-01",
    "specific_dates": [
      "2026-04-20",
      "2026-04-22",
      "2026-04-25"
    ]
  }
}
```

daily

```json

{
  "title": "Daily calls",
  "description": "daily recurrence test",
  "status": "new",
  "recurrence": {
    "type": "daily",
    "start_date": "2026-04-20",
    "end_date": "2026-04-26",
    "every_n_days": 2
  }
}
```

monthly

```json

{
  "title": "Monthly report",
  "description": "monthly recurrence test",
  "status": "new",
  "recurrence": {
    "type": "monthly",
    "start_date": "2026-04-10",
    "end_date": "2026-07-30",
    "day_of_month": 15
  }
}
```
odd_days

```json

{
  "title": "Odd days test",
  "description": "odd recurrence test",
  "status": "new",
  "recurrence": {
    "type": "odd_days",
    "start_date": "2026-04-20",
    "end_date": "2026-04-26"
  }
}
```
even_days
```json

{
  "title": "Even days test",
  "description": "even recurrence test",
  "status": "new",
  "recurrence": {
    "type": "even_days",
    "start_date": "2026-04-20",
    "end_date": "2026-04-26"
  }
}

```

## Как проверить

После запуска API можно открыть Swagger и проверить создание:
- обычной задачи 
- задачи с specific_dates 
- задачи с daily 
- задачи с monthly 
- задачи с odd_days 
- задачи с even_days 
- Результат можно посмотреть через GET /tasks

