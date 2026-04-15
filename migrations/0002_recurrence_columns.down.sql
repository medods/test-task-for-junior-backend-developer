ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_recurrence_start_date_required_check; ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_recurrence_payload_object_check;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_recurrence_type_check;

ALTER TABLE tasks
DROP COLUMN IF EXISTS recurrence_start_date,
DROP COLUMN IF EXISTS recurrence_payload,
DROP COLUMN IF EXISTS recurrence_type;
