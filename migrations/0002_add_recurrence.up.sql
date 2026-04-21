ALTER TABLE tasks
	ADD COLUMN IF NOT EXISTS scheduled_for DATE NOT NULL DEFAULT CURRENT_DATE,
	ADD COLUMN IF NOT EXISTS recurrence_rule_id BIGINT NULL,
	ADD COLUMN IF NOT EXISTS is_modified BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS recurrence_rules (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	active BOOLEAN NOT NULL DEFAULT TRUE,
	schedule_type TEXT NOT NULL,
	starts_on DATE NOT NULL,
	ends_on DATE NULL,
	interval_days INTEGER NULL,
	day_of_month INTEGER NULL,
	day_parity TEXT NULL,
	specific_dates DATE[] NOT NULL DEFAULT '{}',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT chk_recurrence_rules_schedule_type CHECK (
		schedule_type IN ('every_n_days', 'monthly_day', 'specific_dates', 'month_day_parity')
	),
	CONSTRAINT chk_recurrence_rules_interval_days CHECK (interval_days IS NULL OR interval_days > 0),
	CONSTRAINT chk_recurrence_rules_day_of_month CHECK (day_of_month IS NULL OR day_of_month BETWEEN 1 AND 30),
	CONSTRAINT chk_recurrence_rules_day_parity CHECK (day_parity IS NULL OR day_parity IN ('odd', 'even')),
	CONSTRAINT chk_recurrence_rules_date_range CHECK (ends_on IS NULL OR ends_on >= starts_on)
);

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'fk_tasks_recurrence_rule'
	) THEN
		ALTER TABLE tasks
			ADD CONSTRAINT fk_tasks_recurrence_rule
			FOREIGN KEY (recurrence_rule_id)
			REFERENCES recurrence_rules(id)
			ON DELETE SET NULL;
	END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_for ON tasks (scheduled_for DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_rule_id ON tasks (recurrence_rule_id);
CREATE INDEX IF NOT EXISTS idx_recurrence_rules_active ON recurrence_rules (active);

CREATE UNIQUE INDEX IF NOT EXISTS ux_tasks_generated_occurrence
	ON tasks (recurrence_rule_id, scheduled_for)
	WHERE recurrence_rule_id IS NOT NULL;
