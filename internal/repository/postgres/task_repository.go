package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
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
		INSERT INTO tasks (title, description, status, created_at, updated_at, recurrence_type, recurrence_payload, recurrence_start_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, created_at, updated_at, recurrence_type, recurrence_payload, recurrence_start_date
	`

	payload, err := json.Marshal(task.RecurrencePayload)
	if err != nil {
		return nil, err
	}
	var t *time.Time
	if task.RecurrenceStartDate != nil {
		tmp := task.RecurrenceStartDate.Time
		t = &tmp
	}
	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt, task.RecurrenceType, payload, t)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, recurrence_type, recurrence_payload, recurrence_start_date
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
			updated_at = $4,
            recurrence_type = $5,
            recurrence_payload = $6,
            recurrence_start_date = $7
		WHERE id = $8
		RETURNING id, title, description, status, created_at, updated_at, recurrence_type, recurrence_payload, recurrence_start_date
	`

	payload, err := json.Marshal(task.RecurrencePayload)
	if err != nil {
		return nil, err
	}
	var t *time.Time
	if task.RecurrenceStartDate != nil {
		tmp := task.RecurrenceStartDate.Time
		t = &tmp
	}
	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.RecurrenceType, payload, t, task.ID)

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
		SELECT id, title, description, status, created_at, updated_at, recurrence_type, recurrence_payload, recurrence_start_date
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

func (r *Repository) ListByDate(ctx context.Context, date time.Time) ([]taskdomain.Task, error) {
	const query = `
    SELECT id, title, description, status, created_at, updated_at, recurrence_type, recurrence_payload, recurrence_start_date
    FROM tasks
    WHERE
      (recurrence_type = 'none' AND recurrence_start_date = $1::date)
      OR
      (
        recurrence_type = 'daily_every_n'
        AND recurrence_start_date IS NOT NULL
        AND $1::date >= recurrence_start_date
        AND (recurrence_payload->>'interval_days') IS NOT NULL
        AND ((recurrence_payload->>'interval_days')::int > 0)
        AND ((($1::date - recurrence_start_date) % ((recurrence_payload->>'interval_days')::int)) = 0)
      )
      OR
      (
        recurrence_type = 'monthly_day'
        AND (recurrence_payload->>'month_day') IS NOT NULL
        AND ((recurrence_payload->>'month_day')::int BETWEEN 1 AND 30)
        AND date_part('day', $1::date)::int = (recurrence_payload->>'month_day')::int
        AND $1::date >= recurrence_start_date
      )
      OR
      (
        recurrence_type = 'specific_dates'
        AND (recurrence_payload->'specific_dates') ? to_char($1::date, 'YYYY-MM-DD')
      )
      OR
      (
        recurrence_type = 'month_parity'
        AND (recurrence_payload->>'parity') IN ('even','odd')
        AND (
          ((recurrence_payload->>'parity') = 'even' AND (date_part('day', $1::date)::int % 2) = 0)
          OR
          ((recurrence_payload->>'parity') = 'odd' AND (date_part('day', $1::date)::int % 2) = 1)
        )
        AND $1::date >= recurrence_start_date
      );
    `

	normalized := common.NewDate(date.UTC()).Time

	rows, err := r.pool.Query(ctx, query, normalized)
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
		task   taskdomain.Task
		status string
	)

	var payload []byte
	var t *time.Time
	var recType string
	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&recType,
		&payload,
		&t,
	); err != nil {
		return nil, err
	}

	task.RecurrenceType = recurrence.RecurrenceType(recType)

	task.RecurrencePayload = recurrence.RecurrencePayload{
		IntervalDays:  1,
		MonthDay:      0,
		SpecificDates: nil,
		Parity:        recurrence.NoneParity,
	}

	if len(payload) != 0 {
		if err := json.Unmarshal(payload, &task.RecurrencePayload); err != nil {
			return nil, err
		}
	}

	if t != nil {
		d := common.NewDate(*t)
		task.RecurrenceStartDate = &d
	} else {
		task.RecurrenceStartDate = nil
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
