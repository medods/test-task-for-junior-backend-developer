CREATE TABLE task_recurrences (
	id                 BIGSERIAL PRIMARY KEY,
	type               TEXT  NOT NULL CHECK (type IN ('daily', 'monthly', 'specific_dates', 'weekly')),
	config             JSONB NOT NULL,
	next_occurrence_at DATE  NOT NULL
);

ALTER TABLE tasks
	ADD COLUMN recurrence_id BIGINT REFERENCES task_recurrences(id) ON DELETE SET NULL;
