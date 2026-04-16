package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCanceled   Status = "canceled"
)

type Task struct {
	ID                int64
	Title             string
	Description       string
	Status            Status
	ScheduledAt       *time.Time
	RecurrenceGroupID *string
	Recurrence        *RecurrenceConfig
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone, StatusCanceled:
		return true
	default:
		return false
	}
}
