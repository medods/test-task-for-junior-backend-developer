-- Добавляет поле scheduled_for для экземпляров задач
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS scheduled_for TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_for ON tasks (scheduled_for);