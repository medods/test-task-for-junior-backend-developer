ALTER TABLE tasks
ADD COLUMN recurrence_type TEXT NOT NULL DEFAULT 'none',
ADD COLUMN recurrence_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN recurrence_start_date DATE;

ALTER TABLE tasks
ADD CONSTRAINT tasks_recurrence_type_check
CHECK (recurrence_type IN (
  'none',
  'daily_every_n',
  'monthly_day',
  'specific_dates',
  'month_parity'
));

ALTER TABLE tasks
ADD CONSTRAINT tasks_recurrence_payload_object_check
CHECK (jsonb_typeof(recurrence_payload) = 'object');

ALTER TABLE tasks
ADD CONSTRAINT tasks_recurrence_start_date_required_check
CHECK (
  recurrence_type = 'none'
  OR recurrence_start_date IS NOT NULL
);
