-- Добавляет поле last_run_at для отслеживания последнего запуска воркера
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS last_run_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_tasks_last_run_at ON tasks (last_run_at);