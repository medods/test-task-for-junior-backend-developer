package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
		INSERT INTO tasks (
			title, description, status,
			recurrence_type, recurrence_interval, recurrence_day_of_month,
			recurrence_specific_dates, recurrence_is_even, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, title, description, status, 
			recurrence_type, recurrence_interval, recurrence_day_of_month,
			recurrence_specific_dates, recurrence_is_even,
			created_at, updated_at
		`

	var recurrenceJSON *string
	if task.Recurrence != nil && len(task.Recurrence.SpecificDates) > 0 {
		data, err := json.Marshal(task.Recurrence.SpecificDates)
		if err != nil {
			return nil, err
		}
		jsonStr := string(data)
		recurrenceJSON = &jsonStr
	}

	row := r.pool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		nullIfNil(task.Recurrence, "Type"),
		nullIfNil(task.Recurrence, "Interval"),
		nullIfNil(task.Recurrence, "DayOfMonth"),
		recurrenceJSON,
		nullIfNil(task.Recurrence, "IsEven"),
		task.CreatedAt,
		task.UpdatedAt,
	)

	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status,
			recurrence_type, recurrence_interval, recurrence_day_of_month,
			recurrence_specific_dates, recurrence_is_even,created_at, updated_at
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
			recurrence_type = $4,
			recurrence_interval = $5,
			recurrence_day_of_month = $6,
			recurrence_specific_dates = $7,
			recurrence_is_even = $8,
			updated_at = $9
		WHERE id = $10
		RETURNING id, title, description, status, 
			recurrence_type, recurrence_interval, recurrence_day_of_month, 
			recurrence_specific_dates, recurrence_is_even, 
			created_at, updated_at	
	`

	var recurrenceJSON *string
	if task.Recurrence != nil && len(task.Recurrence.SpecificDates) > 0 {
		data, err := json.Marshal(task.Recurrence.SpecificDates)
		if err != nil {
			return nil, err
		}
		jsonStr := string(data)
		recurrenceJSON = &jsonStr
	}
	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status,
		nullIfNil(task.Recurrence, "Type"),
		nullIfNil(task.Recurrence, "Interval"),
		nullIfNil(task.Recurrence, "DayOfMonth"),
		recurrenceJSON,
		nullIfNil(task.Recurrence, "IsEven"),
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
func nullIfNil(rec *taskdomain.Recurrence, field string) any {
	if rec == nil {
		return nil
	}

	switch field {
	case "Type":
		return string(rec.Type)
	case "Interval":
		return rec.Interval
	case "DayOfMonth":
		return rec.DayOfMonth
	case "IsEven":
		return rec.IsEven
	default:
		return nil
	}
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status,
			recurrence_type, recurrence_interval, recurrence_day_of_month,
			recurrence_specific_dates, recurrence_is_even,
			created_at, updated_at
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
		task   taskdomain.Task
		status string

		recurrenceType          *string
		recurrenceInterval      *int
		recurrenceDayOfMonth    *int
		recurrenceSpecificDates *string
		recurrenceIsEven        *bool
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&recurrenceType,
		&recurrenceInterval,
		&recurrenceDayOfMonth,
		&recurrenceSpecificDates,
		&recurrenceIsEven,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	if recurrenceType != nil {
		task.Recurrence = &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceType(*recurrenceType),
			Interval:   recurrenceInterval,
			DayOfMonth: recurrenceDayOfMonth,
			IsEven:     recurrenceIsEven,
		}

		if recurrenceSpecificDates != nil && *recurrenceSpecificDates != "" {
			var dates []time.Time
			if err := json.Unmarshal([]byte(*recurrenceSpecificDates), &dates); err != nil {
				return nil, fmt.Errorf("failed to unmarshal specific_dates: %w", err)
			}
			task.Recurrence.SpecificDates = dates
		}
	}

	return &task, nil
}
