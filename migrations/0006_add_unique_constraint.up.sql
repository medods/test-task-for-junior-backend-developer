-- Добавляет уникальный индекс на (parent_id, scheduled_for) для предотвращения дубликатов экземпляров
ALTER TABLE tasks ADD CONSTRAINT unique_parent_scheduled_for UNIQUE (parent_id, scheduled_for);