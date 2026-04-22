package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// DTO для создания/обновления задачи, теперь поддерживает Periodicity
type taskMutationDTO struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Status      taskdomain.Status       `json:"status"`
	Periodicity *taskdomain.Periodicity `json:"periodicity,omitempty"`
}

// DTO для возврата задачи, теперь поддерживает Periodicity
type taskDTO struct {
	ID          int64                   `json:"id"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Status      taskdomain.Status       `json:"status"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Periodicity *taskdomain.Periodicity `json:"periodicity,omitempty"`
}

// Преобразование Task -> taskDTO с поддержкой Periodicity
func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Periodicity: task.Periodicity,
	}
}
