package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceSQLParams struct {
	recurrenceType       pgtype.Text
	recurrenceInterval   pgtype.Int4
	recurrenceDayOfMonth pgtype.Int4
	recurrenceStartDate  pgtype.Date
	recurrenceDates      pgtype.Array[pgtype.Date]
}

func buildTaskInsertArgs(task *taskdomain.Task) []any {
	params := recurrenceParams(task.Recurrence)

	return []any{
		task.Title,
		task.Description,
		string(task.Status),
		string(task.Mode),
		nullableDateParam(task.ScheduledFor),
		nullableInt64Param(task.SourceTaskID),
		params.recurrenceType,
		params.recurrenceInterval,
		params.recurrenceDayOfMonth,
		params.recurrenceStartDate,
		params.recurrenceDates,
		task.CreatedAt,
		task.UpdatedAt,
	}
}

func buildTaskUpdateArgs(task *taskdomain.Task) []any {
	params := recurrenceParams(task.Recurrence)

	return []any{
		task.Title,
		task.Description,
		string(task.Status),
		string(task.Mode),
		nullableDateParam(task.ScheduledFor),
		nullableInt64Param(task.SourceTaskID),
		params.recurrenceType,
		params.recurrenceInterval,
		params.recurrenceDayOfMonth,
		params.recurrenceStartDate,
		params.recurrenceDates,
		task.UpdatedAt,
		task.ID,
	}
}

func nullableDateParam(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}

	return pgtype.Date{
		Time:  dateOnlyUTC(*value),
		Valid: true,
	}
}

func nullableInt64Param(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}

	return pgtype.Int8{
		Int64: *value,
		Valid: true,
	}
}

func recurrenceParams(recurrence *taskdomain.Recurrence) recurrenceSQLParams {
	if recurrence == nil {
		return recurrenceSQLParams{}
	}

	params := recurrenceSQLParams{
		recurrenceType: pgtype.Text{
			String: string(recurrence.Type),
			Valid:  true,
		},
		recurrenceStartDate: nullableDateParam(recurrence.StartDate),
		recurrenceDates:     dateArrayParam(recurrence.Dates),
	}

	if recurrence.EveryNDays != 0 {
		params.recurrenceInterval = pgtype.Int4{
			Int32: int32(recurrence.EveryNDays),
			Valid: true,
		}
	}

	if recurrence.DayOfMonth != 0 {
		params.recurrenceDayOfMonth = pgtype.Int4{
			Int32: int32(recurrence.DayOfMonth),
			Valid: true,
		}
	}

	return params
}

func dateArrayParam(dates []time.Time) pgtype.Array[pgtype.Date] {
	if len(dates) == 0 {
		return pgtype.Array[pgtype.Date]{}
	}

	elements := make([]pgtype.Date, 0, len(dates))
	for _, date := range dates {
		elements = append(elements, pgtype.Date{
			Time:  dateOnlyUTC(date),
			Valid: true,
		})
	}

	return pgtype.Array[pgtype.Date]{
		Elements: elements,
		Dims: []pgtype.ArrayDimension{
			{
				Length:     int32(len(elements)),
				LowerBound: 1,
			},
		},
		Valid: true,
	}
}
