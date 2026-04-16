package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type ListFilter struct {
	RecurrenceGroupID *string
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt *time.Time
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt *time.Time
}

type CreateRecurringInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Recurrence  taskdomain.RecurrenceInput
}

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	CreateBatch(ctx context.Context, tasks []*taskdomain.Task) ([]taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	DeleteByGroupID(ctx context.Context, groupID string) (int64, error)
	List(ctx context.Context, filter ListFilter) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter ListFilter) ([]taskdomain.Task, error)
	CreateRecurring(ctx context.Context, input CreateRecurringInput) ([]taskdomain.Task, error)
	DeleteByGroupID(ctx context.Context, groupID string) (int64, error)
}
