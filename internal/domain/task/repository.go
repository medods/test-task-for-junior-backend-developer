package task

import (
	"context"
	"time"
)

type TaskRepository interface {
	Create(ctx context.Context, t *Task) (*Task, error)
	GetByID(ctx context.Context, id int64) (*Task, error)
	List(ctx context.Context, filter TaskFilter) ([]*Task, error)
	Update(ctx context.Context, t *Task) (*Task, error)
	Delete(ctx context.Context, id int64) error

	CreateRecurringWithInstances(ctx context.Context, tmpl *Task, cfg *RecurrenceConfig) (*Task, error)

	InsertInstancesIdempotent(ctx context.Context, parentID int64, dates []time.Time) (int, error)

	GetActiveTemplatesByDate(ctx context.Context, asOf time.Time, limit, instanceID, clusterSize int) ([]*Task, error)

	GetTaskWithRecurrence(ctx context.Context, id int64) (*Task, *RecurrenceConfig, error)
}