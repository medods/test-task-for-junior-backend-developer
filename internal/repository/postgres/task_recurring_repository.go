package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func (r *Repository) CreateRecurringTemplate(
	ctx context.Context,
	template *taskdomain.Task,
	scheduledDates []time.Time,
) (*taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created, err := insertRecurringTemplateTx(ctx, tx, template)
	if err != nil {
		return nil, err
	}

	if _, err := insertRecurringInstancesTx(ctx, tx, created, scheduledDates); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) UpdateRecurringTemplate(
	ctx context.Context,
	template *taskdomain.Task,
	regenerateFrom *time.Time,
	scheduledDates []time.Time,
) (*taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	updated, err := updateRecurringTemplateTx(ctx, tx, template)
	if err != nil {
		return nil, err
	}

	if regenerateFrom != nil {
		if err := deleteNewRecurringInstancesFromDateTx(ctx, tx, updated.ID, *regenerateFrom); err != nil {
			return nil, err
		}
	} else {
		if err := deleteNewRecurringInstancesTx(ctx, tx, updated.ID); err != nil {
			return nil, err
		}
	}

	if _, err := insertRecurringInstancesTx(ctx, tx, updated, scheduledDates); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *Repository) DeleteRecurringTemplate(ctx context.Context, templateID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := deleteNewRecurringInstancesTx(ctx, tx, templateID); err != nil {
		return err
	}

	if err := deleteRecurringTemplateTx(ctx, tx, templateID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Repository) ListRecurringTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
	SELECT ` + taskColumns + `
	FROM tasks
	WHERE mode = 'recurring_template'
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

func insertRecurringTemplateTx(
	ctx context.Context,
	tx pgx.Tx,
	template *taskdomain.Task,
) (*taskdomain.Task, error) {
	params := recurrenceParams(template.Recurrence)
	const query = `
	INSERT INTO tasks (
		title,
		description,
		status,
		mode,
		scheduled_for,
		source_task_id,
		recurrence_type,
		recurrence_interval,
		recurrence_day_of_month,
		recurrence_start_date,
		recurrence_dates,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, NULL, NULL, $5, $6, $7, $8, $9, $10, $11)
	RETURNING ` + taskColumns
	row := tx.QueryRow(
		ctx,
		query,
		template.Title,
		template.Description,
		string(template.Status),
		string(taskdomain.ModeRecurringTemplate),
		params.recurrenceType,
		params.recurrenceInterval,
		params.recurrenceDayOfMonth,
		params.recurrenceStartDate,
		params.recurrenceDates,
		template.CreatedAt,
		template.UpdatedAt,
	)

	return scanTask(row)
}

func updateRecurringTemplateTx(
	ctx context.Context,
	tx pgx.Tx,
	template *taskdomain.Task,
) (*taskdomain.Task, error) {
	params := recurrenceParams(template.Recurrence)
	const query = `
	UPDATE tasks
	SET title = $1,
		description = $2,
		status = $3,
		mode = $4,
		scheduled_for = NULL,
		source_task_id = NULL,
		recurrence_type = $5,
		recurrence_interval = $6,
		recurrence_day_of_month = $7,
		recurrence_start_date = $8,
		recurrence_dates = $9,
		updated_at = $10
	WHERE id = $11
		AND mode = 'recurring_template'
	RETURNING ` + taskColumns

	row := tx.QueryRow(
		ctx,
		query,
		template.Title,
		template.Description,
		string(template.Status),
		string(taskdomain.ModeRecurringTemplate),
		params.recurrenceType,
		params.recurrenceInterval,
		params.recurrenceDayOfMonth,
		params.recurrenceStartDate,
		params.recurrenceDates,
		template.UpdatedAt,
		template.ID,
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

func deleteRecurringTemplateTx(
	ctx context.Context,
	tx pgx.Tx,
	templateID int64,
) error {
	const query = `
	DELETE FROM tasks
	WHERE id = $1
		AND mode = 'recurring_template'
`
	result, err := tx.Exec(ctx, query, templateID)
	if err != nil {
		if isForeignKeyViolation(err) {
			return taskdomain.ErrTaskHasGeneratedInstances
		}

		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
