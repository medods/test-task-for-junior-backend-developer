package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)

	// Специальные методы для работы с экземплярами шаблона
	GetInstancesByTemplateID(ctx context.Context, templateID int64) ([]taskdomain.Task, error)
	UpdateInstance(ctx context.Context, instance *taskdomain.Task) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt time.Time
	Recurrence  *taskdomain.RecurrenceConfig
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt time.Time
	Recurrence  *taskdomain.RecurrenceConfig
}
