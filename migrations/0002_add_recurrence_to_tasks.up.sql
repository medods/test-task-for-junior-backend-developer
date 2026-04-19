-- Добавляем scheduled_at (дата и время выполнения)
ALTER TABLE tasks
    ADD COLUMN scheduled_at TIMESTAMPTZ;

-- Добавляем JSONB поле для хранения настроек повторения
ALTER TABLE tasks
    ADD COLUMN recurrence JSONB;

-- Флаг: является ли задача шаблоном для повторяющихся задач
ALTER TABLE tasks
    ADD COLUMN is_recurring_template BOOLEAN DEFAULT FALSE;

-- Ссылка на родительский шаблон (для созданных экземпляров)
ALTER TABLE tasks
    ADD COLUMN parent_template_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE;

-- Индекс для быстрого поиска экземпляров шаблона
CREATE INDEX idx_tasks_parent_template_id ON tasks(parent_template_id);

-- GIN индекс для эффективного поиска по JSONB
CREATE INDEX idx_tasks_recurrence ON tasks USING GIN (recurrence);

-- Индекс для фильтрации шаблонов
CREATE INDEX idx_tasks_is_recurring_template ON tasks(is_recurring_template) WHERE is_recurring_template = true;