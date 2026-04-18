package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceDTO struct {
	Type          string   `json:"type"`
	StartDate     string   `json:"start_date"`
	EndDate       *string  `json:"end_date,omitempty"`
	EveryNDays    *int     `json:"every_n_days,omitempty"`
	DayOfMonth    *int     `json:"day_of_month,omitempty"`
	SpecificDates []string `json:"specific_dates,omitempty"`
}

type taskMutationDTO struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	ScheduledFor *string           `json:"scheduled_for,omitempty"`
	Recurrence   *recurrenceDTO    `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	ScheduledFor *time.Time        `json:"scheduled_for,omitempty"`
	Status       taskdomain.Status `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		ScheduledFor: task.ScheduledFor,
	}
}
