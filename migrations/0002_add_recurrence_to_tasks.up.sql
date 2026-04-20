ALTER TABLE tasks
    ADD COLUMN scheduled_at TIMESTAMPTZ,                                            --дата и время выполнения
    ADD COLUMN recurrence JSONB,                                                    --настройки повторения
    ADD COLUMN is_recurring_template BOOLEAN DEFAULT FALSE,                         --флаг шаблон ли
    ADD COLUMN parent_template_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE;    --ссылка на родительский шаблон

CREATE INDEX idx_tasks_recurrence ON tasks USING GIN (recurrence);                                               --индекс для поиска по JSONB
CREATE INDEX idx_tasks_is_recurring_template ON tasks(is_recurring_template) WHERE is_recurring_template = true; --индекс для фильтрации шаблонов
CREATE INDEX idx_tasks_parent_template_scheduled ON tasks (parent_template_id, scheduled_at);                    --индекс получения экземпляров шаблона