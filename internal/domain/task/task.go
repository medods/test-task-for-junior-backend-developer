package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

type RecurrenceType string

const (
	RecurrenceNone          RecurrenceType = "none"
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenOdd       RecurrenceType = "even_odd"
)

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceNone, RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceEvenOdd:
		return true
	default:
		return false
	}
}

type Recurrence struct {
	Type     RecurrenceType `json:"type"`
	Interval *int           `json:"interval,omitempty"` // daily: каждые N дней
	Day      *int           `json:"day,omitempty"`      // monthly: число месяца 1–30
	Dates    []time.Time    `json:"dates,omitempty"`    // specific_dates
	Even   	 *bool     		`json:"even,omitempty"`   // even_odd
	Start    *time.Time     `json:"start,omitempty"`
	End      *time.Time     `json:"end,omitempty"` // nil = бессрочно
}

type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Recurrence  Recurrence `json:"recurrence"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}