ALTER TABLE tasks
    ADD COLUMN scheduled_date TIMESTAMPTZ,
    ADD COLUMN recurrence_type TEXT,
    ADD COLUMN recurrence_group_id UUID;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_recurrence_group_id ON tasks (recurrence_group_id);
