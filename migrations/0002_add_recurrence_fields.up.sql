-- migrations/0002_add_recurrence_fields.up.sql
ALTER TABLE tasks ADD COLUMN recurrence_type VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN recurrence_every_n INTEGER;
ALTER TABLE tasks ADD COLUMN monthly_day INTEGER CHECK (monthly_day BETWEEN 1 AND 30);
ALTER TABLE tasks ADD COLUMN specific_dates JSONB;
ALTER TABLE tasks ADD COLUMN even_odd VARCHAR(4) CHECK (even_odd IN ('even', 'odd'));

-- Создаем индекс для поиска по типу периодичности
CREATE INDEX idx_tasks_recurrence_type ON tasks(recurrence_type);

-- Комментарии для документации
COMMENT ON COLUMN tasks.recurrence_type IS 'Тип периодичности: daily, monthly_day, specific, even_odd';
COMMENT ON COLUMN tasks.recurrence_every_n IS 'Для daily: интервал в днях';
COMMENT ON COLUMN tasks.monthly_day IS 'Для monthly_day: число месяца от 1 до 30';
COMMENT ON COLUMN tasks.specific_dates IS 'Для specific: JSON массив дат в формате RFC3339';
COMMENT ON COLUMN tasks.even_odd IS 'Для even_odd: even или odd';