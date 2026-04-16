ALTER TABLE tasks ADD COLUMN scheduled_at TIMESTAMPTZ;
ALTER TABLE tasks ADD COLUMN recurrence_group_id TEXT;
ALTER TABLE tasks ADD COLUMN recurrence JSONB;

CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_group_id ON tasks (recurrence_group_id);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_at ON tasks (scheduled_at);
