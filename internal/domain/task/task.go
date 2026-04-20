package task

import (
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID                  int64             `json:"id"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Status              Status            `json:"status"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	ScheduledAt         time.Time         `json:"scheduled_at"`
	Recurrence          *RecurrenceConfig `json:"recurrence,omitempty"`
	IsRecurringTemplate bool              `json:"is_recurring_template"`
	ParentTemplateID    *int64            `json:"parent_template_id,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
