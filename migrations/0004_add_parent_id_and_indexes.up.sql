-- Добавляет поле parent_id для связи экземпляров задач с родительской задачей
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS parent_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE;

-- Индекс по parent_id
CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks (parent_id);

-- Композитный индекс по scheduled_for и parent_id
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_for_parent_id ON tasks (scheduled_for, parent_id);