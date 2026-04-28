package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskCreateDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Mode        taskdomain.Mode   `json:"mode"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskUpdateDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Mode        *taskdomain.Mode  `json:"mode,omitempty"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
}

type recurrenceDTO struct {
	Type       taskdomain.RecurrenceType `json:"type"`
	EveryNDays int                       `json:"every_n_days,omitempty"`
	DayOfMonth int                       `json:"day_of_month,omitempty"`
	StartDate  *string                   `json:"start_date,omitempty"`
	Dates      []string                  `json:"dates,omitempty"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	Mode         taskdomain.Mode   `json:"mode"`
	ScheduledFor *string           `json:"scheduled_for,omitempty"`
	SourceTaskID *int64            `json:"source_task_id,omitempty"`
	Recurrence   *recurrenceDTO    `json:"recurrence,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}
