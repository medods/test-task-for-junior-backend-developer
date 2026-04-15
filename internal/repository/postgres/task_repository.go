package postgres

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	taskusecase "example.com/taskservice/internal/usecase/task"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		INSERT INTO tasks (title, description, status, created_at, updated_at, scheduled_for, schedule_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, created_at, updated_at, scheduled_for, schedule_id
	`

	row := infrastructurepostgres.DBFromContext(ctx, r.pool).QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
		task.ScheduledFor,
		task.ScheduleID,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at, scheduled_for, schedule_id
		FROM tasks
		WHERE id = $1
	`

	row := infrastructurepostgres.DBFromContext(ctx, r.pool).QueryRow(ctx, query, id)
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
		WHERE id = $5
		RETURNING id, title, description, status, created_at, updated_at, scheduled_for, schedule_id
	`

	row := infrastructurepostgres.DBFromContext(ctx, r.pool).QueryRow(ctx, query, task.Title, task.Description, task.Status, task.UpdatedAt, task.ID)
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

	result, err := infrastructurepostgres.DBFromContext(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context, filter taskusecase.ListInput) ([]taskdomain.Task, error) {
	var (
		queryBuilder strings.Builder
		args         []any
		conditions   []string
	)

	queryBuilder.WriteString(`
		SELECT id, title, description, status, created_at, updated_at, scheduled_for, schedule_id
		FROM tasks
	`)

	if filter.DateFrom != nil {
		args = append(args, *filter.DateFrom)
		conditions = append(conditions, "scheduled_for >= $"+parameterIndex(len(args)))
	}

	if filter.DateTo != nil {
		args = append(args, *filter.DateTo)
		conditions = append(conditions, "scheduled_for <= $"+parameterIndex(len(args)))
	}

	if filter.ScheduleID != nil {
		args = append(args, *filter.ScheduleID)
		conditions = append(conditions, "schedule_id = $"+parameterIndex(len(args)))
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	queryBuilder.WriteString(" ORDER BY COALESCE(scheduled_for, created_at::date) DESC, id DESC")

	rows, err := infrastructurepostgres.DBFromContext(ctx, r.pool).Query(ctx, queryBuilder.String(), args...)
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

func (r *Repository) CreateGenerated(ctx context.Context, task *taskdomain.Task) error {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at, scheduled_for, schedule_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT DO NOTHING
	`

	_, err := infrastructurepostgres.DBFromContext(ctx, r.pool).Exec(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
		task.ScheduledFor,
		task.ScheduleID,
	)

	return err
}

func (r *Repository) DeleteFutureNewGenerated(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error {
	const query = `
		DELETE FROM tasks
		WHERE schedule_id = $1
		  AND status = $2
		  AND scheduled_for >= $3
		  AND scheduled_for <= $4
	`

	_, err := infrastructurepostgres.DBFromContext(ctx, r.pool).Exec(ctx, query, scheduleID, taskdomain.StatusNew, from, to)
	return err
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task         taskdomain.Task
		status       string
		scheduledFor pgtype.Date
		scheduleID   pgtype.Int8
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&scheduledFor,
		&scheduleID,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	if scheduledFor.Valid {
		value := scheduledFor.Time.UTC()
		task.ScheduledFor = &value
	}
	if scheduleID.Valid {
		value := scheduleID.Int64
		task.ScheduleID = &value
	}

	return &task, nil
}

func parameterIndex(index int) string {
	return strconv.Itoa(index)
}
