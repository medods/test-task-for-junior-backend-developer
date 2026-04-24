package postgres

import (
	"context"
	"database/sql"
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
	periodPayload, err := marshalPeriod(task.Period)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO tasks (title, description, status, period, period_anchor_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, period, period_anchor_at, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, periodPayload, task.PeriodAnchorAt, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, period, period_anchor_at, created_at, updated_at
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
	periodPayload, err := marshalPeriod(task.Period)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			period = $4,
			period_anchor_at = $5,
			updated_at = $6
		WHERE id = $7
		RETURNING id, title, description, status, period, period_anchor_at, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, periodPayload, task.PeriodAnchorAt, task.UpdatedAt, task.ID)
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
		SELECT id, title, description, status, period, period_anchor_at, created_at, updated_at
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

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task         taskdomain.Task
		status       string
		periodData   []byte
		periodAnchor sql.NullTime
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&periodData,
		&periodAnchor,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	if periodAnchor.Valid {
		value := periodAnchor.Time.UTC()
		task.PeriodAnchorAt = &value
	}

	period, err := unmarshalPeriod(periodData)
	if err != nil {
		return nil, err
	}
	task.Period = period

	return &task, nil
}

func marshalPeriod(period *taskdomain.Period) ([]byte, error) {
	if period == nil {
		return nil, nil
	}

	return json.Marshal(period)
}

func unmarshalPeriod(raw []byte) (*taskdomain.Period, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var period taskdomain.Period
	if err := json.Unmarshal(raw, &period); err != nil {
		return nil, err
	}

	return &period, nil
}
