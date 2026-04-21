package recurrence

import (
	"context"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
)

type Repository interface {
	Create(ctx context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error)
	GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error)
	Update(ctx context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error)
	Delete(ctx context.Context, id int64, fromDate time.Time) error
	List(ctx context.Context) ([]recurrencedomain.Rule, error)
	ListActive(ctx context.Context) ([]recurrencedomain.Rule, error)
	SyncGeneratedTasks(ctx context.Context, rule *recurrencedomain.Rule, fromDate time.Time, occurrences []time.Time) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*recurrencedomain.Rule, error)
	GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*recurrencedomain.Rule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]recurrencedomain.Rule, error)
	MaterializeActive(ctx context.Context) error
}

type CreateInput struct {
	Title         string
	Description   string
	Active        *bool
	ScheduleType  recurrencedomain.ScheduleType
	StartsOn      *time.Time
	EndsOn        *time.Time
	EveryNDays    *int
	DayOfMonth    *int
	DayParity     *recurrencedomain.DayParity
	SpecificDates []time.Time
}

type UpdateInput = CreateInput
