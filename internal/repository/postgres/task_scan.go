package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	taskdomain "example.com/taskservice/internal/domain/task"
)

const taskColumns = `
	id,
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
`

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task                 taskdomain.Task
		status               string
		mode                 string
		scheduledFor         pgtype.Date
		sourceTaskID         pgtype.Int8
		recurrenceType       pgtype.Text
		recurrenceInterval   pgtype.Int4
		recurrenceDayOfMonth pgtype.Int4
		recurrenceStartDate  pgtype.Date
		recurrenceDates      pgtype.Array[pgtype.Date]
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&mode,
		&scheduledFor,
		&sourceTaskID,
		&recurrenceType,
		&recurrenceInterval,
		&recurrenceDayOfMonth,
		&recurrenceStartDate,
		&recurrenceDates,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	recurrence, err := buildRecurrenceFromRow(
		recurrenceType,
		recurrenceInterval,
		recurrenceDayOfMonth,
		recurrenceStartDate,
		recurrenceDates,
	)
	if err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	task.Mode = taskdomain.Mode(mode)
	task.ScheduledFor = scanNullableDate(scheduledFor)
	task.SourceTaskID = scanNullableInt64(sourceTaskID)
	task.Recurrence = recurrence

	return &task, nil
}

func buildRecurrenceFromRow(
	recurrenceType pgtype.Text,
	recurrenceInterval pgtype.Int4,
	recurrenceDayOfMonth pgtype.Int4,
	recurrenceStartDate pgtype.Date,
	recurrenceDates pgtype.Array[pgtype.Date],
) (*taskdomain.Recurrence, error) {
	if !recurrenceType.Valid {
		return nil, nil
	}

	recurrence := &taskdomain.Recurrence{
		Type: taskdomain.RecurrenceType(recurrenceType.String),
	}

	if recurrenceInterval.Valid {
		recurrence.EveryNDays = int(recurrenceInterval.Int32)
	}

	if recurrenceDayOfMonth.Valid {
		recurrence.DayOfMonth = int(recurrenceDayOfMonth.Int32)
	}

	if recurrenceStartDate.Valid {
		startDate := dateOnlyUTC(recurrenceStartDate.Time)
		recurrence.StartDate = &startDate
	}

	if recurrenceDates.Valid {
		recurrence.Dates = make([]time.Time, 0, len(recurrenceDates.Elements))

		for _, value := range recurrenceDates.Elements {
			if !value.Valid {
				return nil, fmt.Errorf("recurrence_dates must not contain null")
			}

			recurrence.Dates = append(recurrence.Dates, dateOnlyUTC(value.Time))
		}
	}

	return recurrence, nil
}

func scanNullableDate(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}

	date := dateOnlyUTC(value.Time)

	return &date
}

func scanNullableInt64(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	result := value.Int64

	return &result
}

func dateOnlyUTC(value time.Time) time.Time {
	year, month, day := value.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
