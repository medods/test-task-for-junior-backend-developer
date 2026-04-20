CREATE TYPE recurrence_type AS ENUM (
    'none',
    'daily',
    'monthly',
    'specific_dates',
    'even_odd'
);

CREATE TABLE IF NOT EXISTS tasks (
    id                   BIGSERIAL       PRIMARY KEY,
    title                TEXT            NOT NULL,
    description          TEXT            NOT NULL DEFAULT '',
    status               TEXT            NOT NULL,
    recurrence_type      recurrence_type NOT NULL DEFAULT 'none',
    recurrence_interval  INTEGER 		 NULL,
    recurrence_day       INTEGER	 	 NULL,
    recurrence_dates     DATE[]		 	 NULL,
    recurrence_even    	 BOOLEAN		 NULL,
    recurrence_start     DATE			 NULL,
    recurrence_end       DATE			 NULL,
    created_at           TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_rec_daily
        CHECK (recurrence_type != 'daily'
            OR (recurrence_interval IS NOT NULL AND recurrence_interval >= 1)),

    CONSTRAINT chk_rec_monthly
        CHECK (recurrence_type != 'monthly'
            OR (recurrence_day IS NOT NULL AND recurrence_day BETWEEN 1 AND 30)),

    CONSTRAINT chk_rec_specific_dates
        CHECK (recurrence_type != 'specific_dates'
            OR (recurrence_dates IS NOT NULL AND cardinality(recurrence_dates) > 0)),

    CONSTRAINT chk_rec_even_odd
        CHECK (recurrence_type != 'even_odd'
            OR recurrence_even IS NOT NULL),

    CONSTRAINT chk_rec_date_range
        CHECK (recurrence_end IS NULL OR recurrence_start <= recurrence_end)
);

CREATE INDEX IF NOT EXISTS idx_tasks_status          ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_type ON tasks (recurrence_type);