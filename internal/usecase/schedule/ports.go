package schedule

import (
	"context"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

type Repository interface {
	Create(ctx context.Context, schedule *scheduledomain.TaskSchedule) (*scheduledomain.TaskSchedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.TaskSchedule, error)
	List(ctx context.Context) ([]scheduledomain.TaskSchedule, error)
	Update(ctx context.Context, schedule *scheduledomain.TaskSchedule) (*scheduledomain.TaskSchedule, error)
	Deactivate(ctx context.Context, id int64, updatedAt time.Time) error
	ListActiveBetween(ctx context.Context, from time.Time, to time.Time) ([]scheduledomain.TaskSchedule, error)
}

type WindowMaterializer interface {
	MaterializeScheduleWindow(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error
	RebuildScheduleWindow(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*scheduledomain.TaskSchedule, error)
	GetByID(ctx context.Context, id int64) (*scheduledomain.TaskSchedule, error)
	List(ctx context.Context) ([]scheduledomain.TaskSchedule, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*scheduledomain.TaskSchedule, error)
	Deactivate(ctx context.Context, id int64) error
}

type CreateInput struct {
	Title       string
	Description string
	StartDate   time.Time
	EndDate     *time.Time
	Active      *bool
	Rule        scheduledomain.RecurrenceRule
}

type UpdateInput struct {
	Title       string
	Description string
	StartDate   time.Time
	EndDate     *time.Time
	Active      *bool
	Rule        scheduledomain.RecurrenceRule
}
