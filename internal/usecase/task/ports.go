package task

import (
	"context"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type CreateInput struct {
	Title            string
	Description      string
	Status           taskdomain.Status
	RecurrenceType   taskdomain.RecurrenceType `json:"recurrence_type"`
	RecurrenceEveryN *int                      `json:"recurrence_every_n"`
	MonthlyDay       *int                      `json:"monthly_day"`
	SpecificDates    taskdomain.TimeSlice      `json:"specific_dates"`
	EvenOdd          *string                   `json:"even_odd"`
}

type UpdateInput struct {
	Title            *string
	Description      *string
	Status           *taskdomain.Status
	RecurrenceType   *taskdomain.RecurrenceType `json:"recurrence_type"`
	RecurrenceEveryN *int                       `json:"recurrence_every_n"`
	MonthlyDay       *int                       `json:"monthly_day"`
	SpecificDates    taskdomain.TimeSlice       `json:"specific_dates"`
	EvenOdd          *string                    `json:"even_odd"`
}
