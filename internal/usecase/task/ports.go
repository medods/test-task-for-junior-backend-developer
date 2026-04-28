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
	CreateRecurringTemplate(ctx context.Context, template *taskdomain.Task, scheduledDates []time.Time) (*taskdomain.Task, error)
	UpdateRecurringTemplate(ctx context.Context, template *taskdomain.Task, regenerateFrom *time.Time, scheduledDates []time.Time) (*taskdomain.Task, error)
	CreateRecurringInstances(ctx context.Context, template *taskdomain.Task, scheduledDates []time.Time) ([]taskdomain.Task, error)
	DeleteRecurringTemplate(ctx context.Context, templateID int64) error
	ListRecurringTemplates(ctx context.Context) ([]taskdomain.Task, error)
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
	Mode        taskdomain.Mode
	Recurrence  *taskdomain.Recurrence
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Mode        *taskdomain.Mode
	Recurrence  *taskdomain.Recurrence
}
