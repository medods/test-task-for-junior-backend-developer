package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceDTO struct {
	Type     string     `json:"type"`
	Interval *int       `json:"interval,omitempty"`
	Day      *int       `json:"day,omitempty"`
	Dates    []time.Time `json:"dates,omitempty"`
	Even     *bool      `json:"even,omitempty"`
	Start    *time.Time `json:"start,omitempty"`
	End      *time.Time `json:"end,omitempty"`
}

type taskMutationDTO struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Recurrence  recurrenceDTO  `json:"recurrence"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  taskdomain.Recurrence `json:"recurrence"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Recurrence:  task.Recurrence,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func recurrenceDTOToDomain(dto recurrenceDTO) taskdomain.Recurrence {
	return taskdomain.Recurrence{
		Type:     taskdomain.RecurrenceType(dto.Type),
		Interval: dto.Interval,
		Day:      dto.Day,
		Dates:    dto.Dates,
		Even:     dto.Even,
		Start:    dto.Start,
		End:      dto.End,
	}
}
