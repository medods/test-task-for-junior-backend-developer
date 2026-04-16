CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_payload_gin ON tasks USING gin (recurrence_payload);
CREATE INDEX IF NOT EXISTS idx_tasks_type_startdate ON tasks (recurrence_type, recurrence_start_date);
