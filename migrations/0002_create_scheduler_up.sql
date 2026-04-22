CREATE TABLE task_templates
(
    id                SERIAL PRIMARY KEY,
    title             TEXT        NOT NULL,
    description       TEXT,
    recurrence_type   VARCHAR(20) NOT NULL,
    recurrence_config JSONB       NOT NULL,
    is_active         BOOLEAN DEFAULT TRUE,
    created_at        TIMESTAMPTZ   NOT NULL,
    updated_at        TIMESTAMPTZ   NOT NULL
);

CREATE INDEX idx_templates_active ON task_templates (id)
    WHERE is_active = true;

ALTER TABLE tasks
    ADD COLUMN task_date TIMESTAMPTZ;

ALTER TABLE tasks
    ADD COLUMN template_id INTEGER REFERENCES task_templates (id);

CREATE UNIQUE INDEX idx_tasks_title_date ON tasks (title, task_date);
CREATE INDEX idx_tasks_date ON tasks (task_date);
