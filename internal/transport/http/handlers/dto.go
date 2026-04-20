package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	Status      taskdomain.Status            `json:"status"`
	ScheduledAt time.Time                    `json:"scheduled_at"`
	Recurrence  *taskdomain.RecurrenceConfig `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	ScheduledAt         time.Time                    `json:"scheduled_at"`
	Recurrence          *taskdomain.RecurrenceConfig `json:"recurrence,omitempty"`
	IsRecurringTemplate bool                         `json:"is_recurring_template,omitempty"`
	ParentTemplateID    *int64                       `json:"parent_template_id,omitempty"`
	InstancesCreated    int                          `json:"instances_created,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		ScheduledAt: task.ScheduledAt,
	}
	if task.Recurrence != nil {
		dto.Recurrence = task.Recurrence
	}
	if task.ParentTemplateID != nil {
		dto.ParentTemplateID = task.ParentTemplateID
	}

	dto.IsRecurringTemplate = task.IsRecurringTemplate

	return dto
}
