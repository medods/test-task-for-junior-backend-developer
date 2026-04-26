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

	GetRecurrence(ctx context.Context, recurrenceID int64) (taskdomain.Recurrence, time.Time, error)
	CreateRecurrence(ctx context.Context, rec taskdomain.Recurrence, nextAt time.Time) (int64, error)
	UpdateNextOccurrence(ctx context.Context, recurrenceID int64, nextAt time.Time) error
	DeleteRecurrence(ctx context.Context, recurrenceID int64) error
	SetTaskRecurrence(ctx context.Context, taskID int64, recurrenceID *int64) error
	ListDueRecurrences(ctx context.Context, today time.Time) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)

	GetRecurrence(ctx context.Context, taskID int64) (int64, taskdomain.Recurrence, time.Time, error)
	SetRecurrence(ctx context.Context, taskID int64, rec taskdomain.Recurrence) error
	DeleteRecurrence(ctx context.Context, taskID int64) error
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}
