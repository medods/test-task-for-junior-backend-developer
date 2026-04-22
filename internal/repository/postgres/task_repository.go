package postgres

import (
	"context"
	"errors"
	taskdomain "example.com/taskservice/internal/domain"
	"example.com/taskservice/internal/transport/http/handlers"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, template_id, task_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (title, task_date) DO NOTHING
		RETURNING id, title, description, status, template_id, task_date, created_at, updated_at
	`

	var templateID *int64
	if task.TemplateID != 0 {
		templateID = &task.TemplateID
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, templateID, task.TaskDate, task.CreatedAt, task.UpdatedAt)

	created, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("task already exists for this date and template")
		}

		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, template_id, task_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, handlers.ErrNotFound
		}

		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			task_date = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, template_id, task_date, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, handlers.ErrNotFound
		}

		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return handlers.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, template_id, task_date, created_at, updated_at
		FROM tasks
		ORDER BY task_date DESC, id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*taskdomain.Task, 0)

	var task *taskdomain.Task

	for rows.Next() {
		task, err = scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get list tasks: %w", err)
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task       taskdomain.Task
		status     string
		templateID *int64
		taskDate   *time.Time
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&templateID,
		&taskDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	task.Status = taskdomain.Status(status)

	if templateID != nil {
		task.TemplateID = *templateID
	}

	if taskDate != nil {
		task.TaskDate = *taskDate
	}

	return &task, nil
}
