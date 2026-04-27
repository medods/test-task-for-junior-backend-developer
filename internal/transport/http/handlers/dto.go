package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Status      taskdomain.Status         `json:"status"`
	Repeat      taskdomain.RepeatInterval `json:"repeat,omitempty"`
	Config      *taskdomain.RepeatConfig  `json:"config,omitempty"`
}

type taskDTO struct {
	ID          int64                     `json:"id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Status      taskdomain.Status         `json:"status"`
	Repeat      taskdomain.RepeatInterval `json:"repeat,omitempty"`
	Config      *taskdomain.RepeatConfig  `json:"config,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Repeat:      task.Repeat,
		Config:      task.Config,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
