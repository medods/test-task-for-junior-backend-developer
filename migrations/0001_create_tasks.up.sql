CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	recurrence_type TEXT,
	recurrence_interval INT,
	recurrence_day_of_month INT,
	recurrence_specific_dates JSONB, 
	recurrence_is_even BOOLEAN, 
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
