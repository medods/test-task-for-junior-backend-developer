package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Recurrence  *Recurrence `json:"recurrence,omitempty"`
}

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceWeekly        RecurrenceType = "weekly"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEven          RecurrenceType = "even"
)

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	Interval      *int           `json:"interval,omitempty"`
	DayOfMonth    *int           `json:"day_of_month,omitempty"`
	SpecificDates []time.Time    `json:"specific_dates,omitempty"`
	IsEven        *bool          `json:"is_even,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
