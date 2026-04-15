CREATE TABLE IF NOT EXISTS task_specific_dates (
  task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  date DATE NOT NULL,
  PRIMARY KEY (task_id, date)
);

CREATE INDEX IF NOT EXISTS idx_task_specific_dates_date ON task_specific_dates (date);

INSERT INTO task_specific_dates (task_id, date)
SELECT t.id, sd.value::date
FROM tasks t
JOIN LATERAL jsonb_array_elements_text(t.recurrence_payload->'specific_dates') sd(value) ON true
WHERE t.recurrence_type = 'specific_dates'
  AND jsonb_typeof(t.recurrence_payload->'specific_dates') = 'array'
ON CONFLICT DO NOTHING;