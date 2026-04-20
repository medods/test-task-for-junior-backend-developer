package postgres

import (
	"context"
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
		INSERT INTO tasks (title, description, status, scheduled_date, recurrence_type, recurrence_group_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, scheduled_date, recurrence_type, recurrence_group_id, created_at, updated_at
		`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.ScheduledDate,
		task.RecurrenceType,
		task.RecurrenceGroupID,
		task.CreatedAt,
		task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) CreateBatch(ctx context.Context, tasks []*taskdomain.Task) ([]*taskdomain.Task, error) {
	if len(tasks) == 0 {
		return nil, nil // можно возвращать ошибку если по условиям линтера nil nil нельзя будет возвращать
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, task := range tasks {
		query := `
		INSERT INTO tasks (title, description, status, scheduled_date, recurrence_type, recurrence_group_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, scheduled_date, recurrence_type, recurrence_group_id, created_at, updated_at
		`
		batch.Queue(
			query,
			task.Title,
			task.Description,
			task.Status,
			task.ScheduledDate,
			task.RecurrenceType,
			task.RecurrenceGroupID,
			task.CreatedAt,
			task.UpdatedAt,
		)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	var created []*taskdomain.Task
	for range tasks {
		row := br.QueryRow()
		t, err := scanTask(row)
		if err != nil {
			return nil, err
		}
		created = append(created, t)
	}

	if err := br.Close(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
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
			scheduled_date = $5
		WHERE id = $6
		RETURNING id, title, description, status, scheduled_date, recurrence_type, recurrence_group_id, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ScheduledDate, task.ID)
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

func (r *Repository) DeleteByGroupID(ctx context.Context, groupID string) error {
	const query = `DELETE FROM tasks WHERE recurrence_group_id = $1`

	result, err := r.pool.Exec(ctx, query, groupID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context, groupID string) ([]taskdomain.Task, error) {
	var (
		rows pgx.Rows
		err  error
		args []any
	)

	query := `
        SELECT id, title, description, status, scheduled_date, recurrence_type, recurrence_group_id, created_at, updated_at
        FROM tasks
    `

	if groupID != "" {
		query += " WHERE recurrence_group_id = $1"
		args = append(args, groupID)
	}

	query += " ORDER BY id DESC"

	rows, err = r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, cErr := scanTask(rows)
		if cErr != nil {
			return nil, cErr
		}

		tasks = append(tasks, *task)
	}

	if cErr := rows.Err(); cErr != nil {
		return nil, cErr
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

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.ScheduledDate,
		&task.RecurrenceType,
		&task.RecurrenceGroupID,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
