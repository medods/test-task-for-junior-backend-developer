package postgres

import (
	"context"
	"encoding/json"
	"errors"
	templatedomain "example.com/taskservice/internal/domain"
	"fmt"
)

var ErrNotFound = errors.New("template not found")

func (r *Repository) CreateTemplate(ctx context.Context, t *templatedomain.Template) (*templatedomain.Template, error) {
	const query = `
        INSERT INTO task_templates (title, description, recurrence_type, recurrence_config, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, title, description, recurrence_type, recurrence_config, is_active, created_at, updated_at
    `

	configRaw, err := json.Marshal(t.RecurrenceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	row := r.pool.QueryRow(ctx, query,
		t.Title,
		t.Description,
		t.RecurrenceType,
		configRaw,
		t.IsActive,
		t.CreatedAt,
		t.UpdatedAt,
	)

	return scanTemplate(row)
}

func (r *Repository) GetActiveTemplates(ctx context.Context) ([]*templatedomain.Template, error) {
	const query = `
        SELECT id, title, description, recurrence_type, recurrence_config, is_active, created_at, updated_at
        FROM task_templates
        WHERE is_active = true
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active templates: %w", err)
	}
	defer rows.Close()

	templates := make([]*templatedomain.Template, 0)

	var template *templatedomain.Template

	for rows.Next() {
		template, err = scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (r *Repository) UpdateTemplateStatus(ctx context.Context, id int64, active bool) error {
	query := `UPDATE task_templates SET is_active = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.pool.Exec(ctx, query, active, id)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {

		return ErrNotFound
	}

	return nil
}

func scanTemplate(scanner taskScanner) (*templatedomain.Template, error) {
	var (
		t         templatedomain.Template
		configRaw []byte
	)

	err := scanner.Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.RecurrenceType,
		&configRaw,
		&t.IsActive,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan template: %w", err)
	}

	if err = json.Unmarshal(configRaw, &t.RecurrenceConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &t, nil
}
