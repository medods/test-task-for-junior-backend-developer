-- Добавляет поле periodicity для хранения настроек повторяемости задачи (Value Object, JSONB)
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS periodicity JSONB;