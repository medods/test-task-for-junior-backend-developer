CREATE FUNCTION tasks_date_array_is_unique_and_not_null(dates DATE[])
RETURNS BOOLEAN
LANGUAGE sql
IMMUTABLE
STRICT
AS $$
    SELECT
        NOT EXISTS (
            SELECT 1
            FROM unnest(dates) AS d(value)
            WHERE d.value IS NULL
        )
        AND cardinality(dates) = (
            SELECT COUNT(DISTINCT d.value)::int
            FROM unnest(dates) AS d(value)
        );
$$;

ALTER TABLE tasks
    ADD COLUMN mode TEXT NOT NULL DEFAULT 'one_time',
    ADD COLUMN scheduled_for DATE NULL,
    ADD COLUMN source_task_id BIGINT NULL,
    ADD COLUMN recurrence_type TEXT NULL,
    ADD COLUMN recurrence_interval INT NULL,
    ADD COLUMN recurrence_day_of_month INT NULL,
    ADD COLUMN recurrence_start_date DATE NULL,
    ADD COLUMN recurrence_dates DATE[] NULL;

ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_source_task_id
        FOREIGN KEY (source_task_id)
        REFERENCES tasks(id)
        ON DELETE RESTRICT,

    ADD CONSTRAINT tasks_mode
        CHECK (mode IN ('one_time', 'recurring_template', 'recurring_instance')),

    ADD CONSTRAINT tasks_recurrence_type
        CHECK (
            recurrence_type IS NULL
            OR recurrence_type IN ('daily', 'monthly', 'specific_dates', 'odd_days', 'even_days')
        ),

    ADD CONSTRAINT tasks_recurrence_interval_positive
        CHECK (recurrence_interval IS NULL OR recurrence_interval > 0),

    ADD CONSTRAINT tasks_recurrence_day_of_month_range
        CHECK (
            recurrence_day_of_month IS NULL
            OR recurrence_day_of_month BETWEEN 1 AND 30
        ),
    ADD CONSTRAINT tasks_source_task_id_not_self
        CHECK (
            source_task_id IS NULL
            OR source_task_id <> id
        ),

    ADD CONSTRAINT tasks_recurrence_dates_valid
        CHECK (
            recurrence_dates IS NULL
            OR (
                array_ndims(recurrence_dates) = 1
                AND cardinality(recurrence_dates) > 0
                AND tasks_date_array_is_unique_and_not_null(recurrence_dates)
            )
        ),

    ADD CONSTRAINT tasks_shape
        CHECK (
            (
                mode = 'one_time'
                AND source_task_id IS NULL
                AND recurrence_type IS NULL
                AND recurrence_interval IS NULL
                AND recurrence_day_of_month IS NULL
                AND recurrence_start_date IS NULL
                AND recurrence_dates IS NULL
            )
            OR
            (
                mode = 'recurring_template'
                AND scheduled_for IS NULL
                AND source_task_id IS NULL
                AND recurrence_type IS NOT NULL
            )
            OR
            (
                mode = 'recurring_instance'
                AND scheduled_for IS NOT NULL
                AND source_task_id IS NOT NULL
                AND recurrence_type IS NULL
                AND recurrence_interval IS NULL
                AND recurrence_day_of_month IS NULL
                AND recurrence_start_date IS NULL
                AND recurrence_dates IS NULL
            )
        ),

    ADD CONSTRAINT tasks_recurring_template_type_shape
        CHECK (
            mode <> 'recurring_template'
            OR (
                CASE recurrence_type
                    WHEN 'daily' THEN
                        recurrence_interval IS NOT NULL
                        AND recurrence_start_date IS NOT NULL
                        AND recurrence_day_of_month IS NULL
                        AND recurrence_dates IS NULL

                    WHEN 'monthly' THEN
                        recurrence_interval IS NULL
                        AND recurrence_day_of_month IS NOT NULL
                        AND recurrence_dates IS NULL

                    WHEN 'specific_dates' THEN
                        recurrence_interval IS NULL
                        AND recurrence_day_of_month IS NULL
                        AND recurrence_start_date IS NULL
                        AND recurrence_dates IS NOT NULL

                    WHEN 'odd_days' THEN
                        recurrence_interval IS NULL
                        AND recurrence_day_of_month IS NULL
                        AND recurrence_dates IS NULL

                    WHEN 'even_days' THEN
                        recurrence_interval IS NULL
                        AND recurrence_day_of_month IS NULL
                        AND recurrence_dates IS NULL

                    ELSE false
                END
            )
        );

CREATE UNIQUE INDEX ux_tasks_recurring_instance_source_task_id_scheduled_for
    ON tasks (source_task_id, scheduled_for)
    WHERE mode = 'recurring_instance';

CREATE INDEX idx_tasks_scheduled_for
    ON tasks (scheduled_for)
    WHERE scheduled_for IS NOT NULL;

CREATE INDEX idx_tasks_source_task_id
    ON tasks (source_task_id)
    WHERE source_task_id IS NOT NULL;