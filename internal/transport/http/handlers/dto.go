package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskCreateDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskUpdateDTO struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	ScheduledDate *time.Time        `json:"scheduled_date,omitempty"`
}

type taskDTO struct {
	ID                int64             `json:"id"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	Status            taskdomain.Status `json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	RecurrenceType    *string           `json:"recurrence_type,omitempty"`
	RecurrenceGroupID *string           `json:"recurrence_group_id,omitempty"`
	ScheduledDate     *time.Time        `json:"scheduled_date,omitempty"`
}

type recurrenceDTO struct {
	Type       string      `json:"type"`
	EveryNDays *int        `json:"every_n_days,omitempty"`
	DayOfMonth *int        `json:"day_of_month,omitempty"`
	Dates      []time.Time `json:"dates,omitempty"`
	Parity     *string     `json:"parity,omitempty"`
	StartDate  time.Time   `json:"start_date"`
	EndDate    time.Time   `json:"end_date"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:                task.ID,
		Title:             task.Title,
		Description:       task.Description,
		Status:            task.Status,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
		RecurrenceType:    task.RecurrenceType,
		RecurrenceGroupID: task.RecurrenceGroupID,
		ScheduledDate:     task.ScheduledDate,
	}
}
