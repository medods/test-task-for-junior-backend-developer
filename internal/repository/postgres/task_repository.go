package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	recurrenceJSON, err := marshalRecurrence(task.Recurrence)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO tasks (title, description, status, scheduled_at, recurrence_group_id, recurrence, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, scheduled_at, recurrence_group_id, recurrence::text, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		task.ScheduledAt, task.RecurrenceGroupID, recurrenceJSON,
		task.CreatedAt, task.UpdatedAt,
	)
	return scanTask(row)
}

func (r *Repository) CreateBatch(ctx context.Context, tasks []*taskdomain.Task) ([]taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const query = `
		INSERT INTO tasks (title, description, status, scheduled_at, recurrence_group_id, recurrence, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, scheduled_at, recurrence_group_id, recurrence::text, created_at, updated_at
	`

	result := make([]taskdomain.Task, 0, len(tasks))
	for _, task := range tasks {
		recurrenceJSON, err := marshalRecurrence(task.Recurrence)
		if err != nil {
			return nil, err
		}

		row := tx.QueryRow(ctx, query,
			task.Title, task.Description, task.Status,
			task.ScheduledAt, task.RecurrenceGroupID, recurrenceJSON,
			task.CreatedAt, task.UpdatedAt,
		)
		created, err := scanTask(row)
		if err != nil {
			return nil, err
		}
		result = append(result, *created)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, scheduled_at, recurrence_group_id, recurrence::text, created_at, updated_at
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
		SET title        = $1,
		    description  = $2,
		    status       = $3,
		    scheduled_at = $4,
		    updated_at   = $5
		WHERE id = $6
		RETURNING id, title, description, status, scheduled_at, recurrence_group_id, recurrence::text, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		task.ScheduledAt, task.UpdatedAt, task.ID,
	)
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

func (r *Repository) DeleteByGroupID(ctx context.Context, groupID string) (int64, error) {
	const query = `DELETE FROM tasks WHERE recurrence_group_id = $1`

	result, err := r.pool.Exec(ctx, query, groupID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (r *Repository) List(ctx context.Context, filter taskusecase.ListFilter) ([]taskdomain.Task, error) {
	query := `
		SELECT id, title, description, status, scheduled_at, recurrence_group_id, recurrence::text, created_at, updated_at
		FROM tasks
	`
	args := []any{}

	if filter.RecurrenceGroupID != nil {
		query += ` WHERE recurrence_group_id = $1`
		args = append(args, *filter.RecurrenceGroupID)
	}

	query += ` ORDER BY id DESC`

	rows, err := r.pool.Query(ctx, query, args...)
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

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task          taskdomain.Task
		status        string
		scheduledAt   pgtype.Timestamptz
		groupID       pgtype.Text
		recurrenceRaw pgtype.Text
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&scheduledAt,
		&groupID,
		&recurrenceRaw,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	if scheduledAt.Valid {
		t := scheduledAt.Time
		task.ScheduledAt = &t
	}
	if groupID.Valid {
		task.RecurrenceGroupID = &groupID.String
	}
	if recurrenceRaw.Valid && recurrenceRaw.String != "" {
		task.Recurrence = new(taskdomain.RecurrenceConfig)
		if err := json.Unmarshal([]byte(recurrenceRaw.String), task.Recurrence); err != nil {
			return nil, err
		}
	}

	return &task, nil
}

func marshalRecurrence(cfg *taskdomain.RecurrenceConfig) ([]byte, error) {
	if cfg == nil {
		return nil, nil
	}
	return json.Marshal(cfg)
}
