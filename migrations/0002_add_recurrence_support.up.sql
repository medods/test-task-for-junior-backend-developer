CREATE TABLE IF NOT EXISTS task_schedules (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	rule_type TEXT NOT NULL,
	rule_payload JSONB NOT NULL,
	start_date DATE NOT NULL,
	end_date DATE NULL,
	active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tasks
	ADD COLUMN IF NOT EXISTS scheduled_for DATE NULL,
	ADD COLUMN IF NOT EXISTS schedule_id BIGINT NULL REFERENCES task_schedules(id);

CREATE INDEX IF NOT EXISTS idx_task_schedules_active_dates
	ON task_schedules (active, start_date, end_date);

CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_for
	ON tasks (scheduled_for);

CREATE INDEX IF NOT EXISTS idx_tasks_schedule_id
	ON tasks (schedule_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tasks_schedule_scheduled_for
	ON tasks (schedule_id, scheduled_for)
	WHERE schedule_id IS NOT NULL;
