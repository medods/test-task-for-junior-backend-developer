package handlers

import (
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ScheduledFor *string           `json:"scheduled_for,omitempty"`
	ScheduleID   *int64            `json:"schedule_id,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	var scheduledFor *string
	if task.ScheduledFor != nil {
		formatted := scheduledomain.FormatDate(*task.ScheduledFor)
		scheduledFor = &formatted
	}

	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		ScheduledFor: scheduledFor,
		ScheduleID:   task.ScheduleID,
	}
}

type recurrenceDTO struct {
	Type       string   `json:"type"`
	EveryDays  *int     `json:"every_days,omitempty"`
	DayOfMonth *int     `json:"day_of_month,omitempty"`
	Dates      []string `json:"dates,omitempty"`
	Parity     string   `json:"parity,omitempty"`
}

type taskScheduleMutationDTO struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	StartDate   string        `json:"start_date"`
	EndDate     *string       `json:"end_date,omitempty"`
	Active      *bool         `json:"active,omitempty"`
	Recurrence  recurrenceDTO `json:"recurrence"`
}

type taskScheduleDTO struct {
	ID          int64         `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	StartDate   string        `json:"start_date"`
	EndDate     *string       `json:"end_date,omitempty"`
	Active      bool          `json:"active"`
	Recurrence  recurrenceDTO `json:"recurrence"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
