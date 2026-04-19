package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at,
		                   scheduled_at, recurrence, is_recurring_template, parent_template_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, title, description, status, created_at, updated_at,
				  scheduled_at, recurrence, is_recurring_template, parent_template_id
	`

	// Сериализация recurrence в JSON, если он есть
	var recurrenceJSON []byte
	if task.Recurrence != nil {
		var err error
		recurrenceJSON, err = json.Marshal(task.Recurrence)
		if err != nil {
			return nil, err
		}
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt,
		task.ScheduledAt, recurrenceJSON, task.IsRecurringTemplate, task.ParentTemplateID)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		       scheduled_at, recurrence, is_recurring_template, parent_template_id
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			updated_at = $4
			scheduled_at = $5,
			recurrence = $6,
			is_recurring_template = $7,
			parent_template_id = $8
		WHERE id = $9
		RETURNING id, title, description, status, created_at, updated_at, 
		    	  scheduled_at, recurrence, is_recurring_template, parent_template_id
	`

	// Сериализация recurrence в JSON, если он есть
	var recurrenceJSON []byte
	if task.Recurrence != nil {
		var err error
		recurrenceJSON, err = json.Marshal(task.Recurrence)
		if err != nil {
			return nil, err
		}
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt,
		task.ScheduledAt, recurrenceJSON, task.IsRecurringTemplate, task.ParentTemplateID, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at,
		       scheduled_at, recurrence, is_recurring_template, parent_template_id
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetInstancesByTemplateID возвращает все экземпляры шаблона
func (r *Repository) GetInstancesByTemplateID(ctx context.Context, templateID int64) ([]taskdomain.Task, error) {
	const query = `
        SELECT id, title, description, status, created_at, updated_at, 
               scheduled_at, recurrence, is_recurring_template, parent_template_id
        FROM tasks
        WHERE parent_template_id = $1
        ORDER BY scheduled_at ASC
    `

	rows, err := r.pool.Query(ctx, query, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instances := make([]taskdomain.Task, 0)
	for rows.Next() {
		instance, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, *instance)
	}

	return instances, nil
}

// UpdateInstance обновляет существующий экземпляр (только изменяемые поля)
func (r *Repository) UpdateInstance(ctx context.Context, instance *taskdomain.Task) error {
	const query = `
        UPDATE tasks
        SET title = $1,
            description = $2,
            status = $3,
            updated_at = $4,
            scheduled_at = $5
        WHERE id = $6 AND parent_template_id IS NOT NULL
    `

	_, err := r.pool.Exec(ctx, query,
		instance.Title,
		instance.Description,
		instance.Status,
		instance.UpdatedAt,
		instance.ScheduledAt,
		instance.ID,
	)
	return err
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task           taskdomain.Task
		status         string
		recurrenceJSON []byte
		parentID       *int64
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.ScheduledAt,
		&recurrenceJSON,
		&task.IsRecurringTemplate,
		&parentID,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.ParentTemplateID = parentID

	//Десериализация recurrence, если он есть
	if len(recurrenceJSON) > 0 {
		var recurrence taskdomain.RecurrenceConfig
		if err := json.Unmarshal(recurrenceJSON, &recurrence); err != nil {
			return nil, err
		}
		task.Recurrence = &recurrence
	}

	return &task, nil
}
