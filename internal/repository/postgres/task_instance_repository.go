package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func (r *Repository) CreateRecurringInstances(
	ctx context.Context,
	template *taskdomain.Task,
	scheduledDates []time.Time,
) ([]taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created, err := insertRecurringInstancesTx(ctx, tx, template, scheduledDates)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return created, nil
}

func deleteNewRecurringInstancesFromDateTx(
	ctx context.Context,
	tx pgx.Tx,
	templateID int64,
	from time.Time,
) error {
	const query = `
	DELETE FROM tasks
	WHERE source_task_id = $1
		AND mode = 'recurring_instance'
		AND status = 'new'
		AND scheduled_for >= $2
`
	_, err := tx.Exec(
		ctx,
		query,
		templateID,
		nullableDateParam(&from),
	)

	return err
}

func deleteNewRecurringInstancesTx(
	ctx context.Context,
	tx pgx.Tx,
	templateID int64,
) error {
	const query = `
	DELETE FROM tasks
	WHERE source_task_id = $1
		AND mode = 'recurring_instance'
		AND status = 'new'
`
	_, err := tx.Exec(ctx, query, templateID)

	return err
}

func insertRecurringInstancesTx(
	ctx context.Context,
	tx pgx.Tx,
	template *taskdomain.Task,
	scheduledDates []time.Time,
) ([]taskdomain.Task, error) {
	if len(scheduledDates) == 0 {
		return []taskdomain.Task{}, nil
	}
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
	SELECT
		$1,
		$2,
		$3,
		'recurring_instance',
		d.scheduled_for,
		$4,
		NULL,
		NULL,
		NULL,
		NULL,
		NULL,
		NOW(),
		NOW()
	FROM unnest($5::date[]) AS d(scheduled_for)
	ON CONFLICT (source_task_id, scheduled_for)
		WHERE mode = 'recurring_instance'
	DO NOTHING
	RETURNING ` + taskColumns

	rows, err := tx.Query(
		ctx,
		query,
		template.Title,
		template.Description,
		string(taskdomain.StatusNew),
		template.ID,
		dateArrayParam(scheduledDates),
	)
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
