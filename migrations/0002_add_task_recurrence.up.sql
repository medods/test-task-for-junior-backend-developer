ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS recurrence JSONB NOT NULL DEFAULT '{"type":"none"}'::jsonb;
