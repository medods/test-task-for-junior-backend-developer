package handlers

import (
	"example.com/taskservice/internal/domain"
	"time"
)

type RecurrenceDTO struct {
	Type   string                  `json:"type"`
	Config domain.RecurrenceConfig `json:"config"`
}

type TaskMutationDTO struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      domain.Status  `json:"status"`
	TaskDate    time.Time      `json:"task_date"`
	Recurrence  *RecurrenceDTO `json:"recurrence,omitempty"`
}

type TaskDTO struct {
	ID          int64         `json:"id"`
	TemplateID  int64         `json:"template_id,omitempty"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      domain.Status `json:"status"`
	TaskDate    time.Time     `json:"task_date"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func newTaskDTO(task *domain.Task) TaskDTO {
	return TaskDTO{
		ID:          task.ID,
		TemplateID:  task.TemplateID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		TaskDate:    task.TaskDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
