package task

import (
	"context"
	"time"

	"example.com/taskservice/internal/domain/common"
	"example.com/taskservice/internal/domain/recurrence"
	task "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *task.Task) (*task.Task, error)
	GetByID(ctx context.Context, id int64) (*task.Task, error)
	Update(ctx context.Context, task *task.Task) (*task.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]task.Task, error)
	ListByDate(ctx context.Context, date time.Time) ([]task.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*task.Task, error)
	GetByID(ctx context.Context, id int64) (*task.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*task.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]task.Task, error)
	ListByDate(ctx context.Context, date time.Time) ([]task.Task, error)
}

type CreateInput struct {
	Title               string
	Description         string
	Status              task.Status
	RecurrenceType      recurrence.RecurrenceType
	RecurrencePayload   recurrence.RecurrencePayload
	RecurrenceStartDate *common.DateType
}

type UpdateInput struct {
	Title               string
	Description         string
	Status              task.Status
	RecurrenceType      recurrence.RecurrenceType
	RecurrencePayload   recurrence.RecurrencePayload
	RecurrenceStartDate *common.DateType
}
