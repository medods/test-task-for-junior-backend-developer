package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		INSERT INTO tasks (
			title, description, status,
			recurrence_type, recurrence_interval, recurrence_day,
			recurrence_dates, recurrence_even,
			recurrence_start, recurrence_end,
			created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			$7, $8,
			$9, $10,
			$11, $12
		)
		RETURNING
			id, title, description, status,
			recurrence_type, recurrence_interval, recurrence_day,
			recurrence_dates, recurrence_even,
			recurrence_start, recurrence_end,
			created_at, updated_at
	`

	args := recurrenceArgs(task)
	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, string(task.Status),
		args[0], args[1], args[2],
		args[3], args[4],
		args[5], args[6],
		task.CreatedAt, task.UpdatedAt,
	)

	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT
			id, title, description, status,
			recurrence_type, recurrence_interval, recurrence_day,
			recurrence_dates, recurrence_even,
			recurrence_start, recurrence_end,
			created_at, updated_at
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
		UPDATE tasks SET
			title                = $1,
			description          = $2,
			status               = $3,
			recurrence_type      = $4,
			recurrence_interval  = $5,
			recurrence_day       = $6,
			recurrence_dates     = $7,
			recurrence_even      = $8,
			recurrence_start     = $9,
			recurrence_end       = $10,
			updated_at           = $11
		WHERE id = $12
		RETURNING
			id, title, description, status,
			recurrence_type, recurrence_interval, recurrence_day,
			recurrence_dates, recurrence_even,
			recurrence_start, recurrence_end,
			created_at, updated_at
	`

	args := recurrenceArgs(task)
	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, string(task.Status),
		args[0], args[1], args[2],
		args[3], args[4],
		args[5], args[6],
		task.UpdatedAt, task.ID,
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

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT 
			id, title, description, status,
			recurrence_type, recurrence_interval, recurrence_day,
			recurrence_dates, recurrence_even,
			recurrence_start, recurrence_end,
			created_at, updated_at
		FROM tasks
		WHERE
			(recurrence_start IS NULL OR recurrence_start <= CURRENT_DATE)
			AND (recurrence_end IS NULL OR recurrence_end >= CURRENT_DATE)
			AND (
				(recurrence_type = 'none'
					AND created_at::date = CURRENT_DATE)

				OR (recurrence_type = 'daily'
					AND recurrence_start IS NOT NULL
					AND (CURRENT_DATE - recurrence_start) % recurrence_interval = 0)

				OR (recurrence_type = 'monthly'
					AND EXTRACT(DAY FROM CURRENT_DATE) = recurrence_day)

				OR (recurrence_type = 'specific_dates'
					AND CURRENT_DATE = ANY(recurrence_dates))

				OR (recurrence_type = 'even_odd'
					AND (
						(recurrence_even = true AND EXTRACT(DAY FROM CURRENT_DATE)::int % 2 = 0)
						OR
						(recurrence_even = false  AND EXTRACT(DAY FROM CURRENT_DATE)::int % 2 = 1)
					))
			)
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

func recurrenceArgs(task *taskdomain.Task) [7]any {
	rec := task.Recurrence

	return [7]any{
		string(rec.Type),
		rec.Interval,
		rec.Day,
		timesToDateArray(rec.Dates),
		rec.Even,
		rec.Start,
		rec.End,
	}
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
		recurrenceType     string
		recurrenceInterval *int
		recurrenceDay      *int
		recurrenceDates    pgtype.Array[pgtype.Date]
		recurrenceEven     *bool
		recurrenceStart    *time.Time
		recurrenceEnd      *time.Time
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&recurrenceType,
		&recurrenceInterval,
		&recurrenceDay,
		&recurrenceDates,
		&recurrenceEven,
		&recurrenceStart,
		&recurrenceEnd,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.Recurrence = taskdomain.Recurrence{
		Type:     taskdomain.RecurrenceType(recurrenceType),
		Interval: recurrenceInterval,
		Day:      recurrenceDay,
		Even:     recurrenceEven,
		Dates:    dateArrayToTimes(recurrenceDates),
		Start:    recurrenceStart,
		End:      recurrenceEnd,
	}

	return &task, nil
}

func timesToDateArray(dates []time.Time) pgtype.Array[pgtype.Date] {
	if len(dates) == 0 {
		return pgtype.Array[pgtype.Date]{}
	}

	elements := make([]pgtype.Date, len(dates))
	for i, d := range dates {
		elements[i] = pgtype.Date{Time: d, Valid: true}
	}

	return pgtype.Array[pgtype.Date]{
		Elements: elements,
		Dims:     []pgtype.ArrayDimension{{Length: int32(len(elements)), LowerBound: 1}},
		Valid:    true,
	}
}

func dateArrayToTimes(arr pgtype.Array[pgtype.Date]) []time.Time {
	if !arr.Valid || len(arr.Elements) == 0 {
		return nil
	}

	dates := make([]time.Time, 0, len(arr.Elements))
	for _, d := range arr.Elements {
		if d.Valid {
			dates = append(dates, d.Time)
		}
	}

	return dates
}