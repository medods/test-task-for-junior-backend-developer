package materializer

import (
	"context"
	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type ScheduleRepository interface {
	GetByID(ctx context.Context, id int64) (*scheduledomain.TaskSchedule, error)
	ListActiveBetween(ctx context.Context, from time.Time, to time.Time) ([]scheduledomain.TaskSchedule, error)
}

type TaskRepository interface {
	CreateGenerated(ctx context.Context, task *taskdomain.Task) error
	DeleteFutureNewGenerated(ctx context.Context, scheduleID int64, from time.Time, to time.Time) error
}
