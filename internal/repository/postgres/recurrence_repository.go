package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func (r *Repository) GetRecurrence(ctx context.Context, recurrenceID int64) (taskdomain.Recurrence, time.Time, error) {
	const query = `
		SELECT type, config, next_occurrence_at
		FROM task_recurrences
		WHERE id = $1
	`

	var (
		typ       string
		config    []byte
		nextAt    time.Time
	)

	if err := r.pool.QueryRow(ctx, query, recurrenceID).Scan(&typ, &config, &nextAt); err != nil {
		return nil, time.Time{}, err
	}

	rec, err := parseRecurrence(typ, config)
	if err != nil {
		return nil, time.Time{}, err
	}

	return rec, nextAt, nil
}

func (r *Repository) CreateRecurrence(ctx context.Context, rec taskdomain.Recurrence, nextAt time.Time) (int64, error) {
	config, err := json.Marshal(rec)
	if err != nil {
		return 0, fmt.Errorf("marshal recurrence config: %w", err)
	}

	const query = `
		INSERT INTO task_recurrences (type, config, next_occurrence_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	if err := r.pool.QueryRow(ctx, query, rec.RecurrenceType(), config, nextAt).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateNextOccurrence(ctx context.Context, recurrenceID int64, nextAt time.Time) error {
	const query = `UPDATE task_recurrences SET next_occurrence_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, nextAt, recurrenceID)
	return err
}

func (r *Repository) DeleteRecurrence(ctx context.Context, recurrenceID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE tasks SET recurrence_id = NULL WHERE recurrence_id = $1`, recurrenceID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM task_recurrences WHERE id = $1`, recurrenceID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) SetTaskRecurrence(ctx context.Context, taskID int64, recurrenceID *int64) error {
	const query = `UPDATE tasks SET recurrence_id = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, recurrenceID, taskID)
	return err
}

func (r *Repository) ListDueRecurrences(ctx context.Context, today time.Time) ([]taskdomain.Task, error) {
	const query = `
		SELECT t.id, t.title, t.description, t.status, t.created_at, t.updated_at, t.recurrence_id
		FROM tasks t
		JOIN task_recurrences tr ON tr.id = t.recurrence_id
		WHERE t.status = 'done'
		  AND tr.next_occurrence_at <= $1
	`

	rows, err := r.pool.Query(ctx, query, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, rows.Err()
}

func parseRecurrence(typ string, config []byte) (taskdomain.Recurrence, error) {
	switch typ {
	case "daily":
		var r taskdomain.DailyRecurrence
		if err := json.Unmarshal(config, &r); err != nil {
			return nil, err
		}
		return r, nil
	case "monthly":
		var r taskdomain.MonthlyRecurrence
		if err := json.Unmarshal(config, &r); err != nil {
			return nil, err
		}
		return r, nil
	case "weekly":
		var r taskdomain.WeeklyRecurrence
		if err := json.Unmarshal(config, &r); err != nil {
			return nil, err
		}
		return r, nil
	case "specific_dates":
		var r taskdomain.SpecificDatesRecurrence
		if err := json.Unmarshal(config, &r); err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, fmt.Errorf("unknown recurrence type: %s", typ)
	}
}
