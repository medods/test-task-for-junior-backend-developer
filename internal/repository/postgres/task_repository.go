package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"log"
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
	       INSERT INTO tasks (title, description, status, created_at, updated_at, periodicity, scheduled_for, parent_id, last_run_at)
	       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	       RETURNING id, title, description, status, created_at, updated_at, periodicity, scheduled_for, parent_id, last_run_at
       `

	var periodicityJSON any
	if task.Periodicity != nil {
		b, err := json.Marshal(task.Periodicity)
		if err != nil {
			return nil, err
		}
		periodicityJSON = b
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt, periodicityJSON, task.ScheduledFor, task.ParentID, task.LastRunAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
	       SELECT id, title, description, status, created_at, updated_at, periodicity, scheduled_for, parent_id, last_run_at
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

// FindByTemplateAndDate ищет задачу-экземпляр по шаблону (parent_id) и дате
func (r *Repository) FindByTemplateAndDate(ctx context.Context, templateID int64, date string) (*taskdomain.Task, error) {
	const query = `
	       SELECT id, title, description, status, created_at, updated_at, periodicity, scheduled_for, parent_id, last_run_at
	       FROM tasks
	       WHERE scheduled_for = $1 AND parent_id = $2
       `
	row := r.pool.QueryRow(ctx, query, date, templateID)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
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
			updated_at = $4,
			periodicity = $5,
			last_run_at = $6
		WHERE id = $7
		RETURNING id, title, description, status, created_at, updated_at, periodicity, scheduled_for, parent_id, last_run_at
	`

	var periodicityJSON any
	if task.Periodicity != nil {
		b, err := json.Marshal(task.Periodicity)
		if err != nil {
			return nil, err
		}
		periodicityJSON = b
	}

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, periodicityJSON, task.LastRunAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) UpdateLastRunAt(ctx context.Context, id int64, lastRunAt *time.Time) error {
   const query = `UPDATE tasks SET last_run_at = $1, updated_at = NOW() WHERE id = $2`
    log.Printf("[REPO] Updating last_run_at for task %d to %v", id, lastRunAt)
    res, err := r.pool.Exec(ctx, query, lastRunAt, id)
    if err != nil {
        log.Printf("[REPO] UpdateLastRunAt error: %v", err)
        return err
    }
    log.Printf("[REPO] Rows affected: %d", res.RowsAffected())
    return nil
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
	       SELECT id, title, description, status, created_at, updated_at, periodicity, scheduled_for, parent_id, last_run_at
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

func (r *Repository) ListPeriodic(ctx context.Context) ([]*taskdomain.Task, error) {
	const query = `
	       SELECT id, title, description, status, created_at, updated_at, periodicity, scheduled_for, parent_id, last_run_at
	       FROM tasks
	       WHERE periodicity IS NOT NULL
	       ORDER BY id DESC
       `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task            taskdomain.Task
		status          string
		periodicityJSON *string
		scheduledFor    *time.Time
		parentID        *int64
		lastRunAt       *time.Time
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&periodicityJSON,
		&scheduledFor,
		&parentID,
		&lastRunAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	if periodicityJSON != nil {
		var p taskdomain.Periodicity
		if err := json.Unmarshal([]byte(*periodicityJSON), &p); err == nil {
			task.Periodicity = &p
		}
	}
	task.ScheduledFor = scheduledFor
	task.ParentID = parentID
	task.LastRunAt = lastRunAt

	return &task, nil
}
