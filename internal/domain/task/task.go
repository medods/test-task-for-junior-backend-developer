package task

import "time"

type Status string
type Mode string
type RecurrenceType string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

const (
	ModeOneTime           Mode = "one_time"
	ModeRecurringTemplate Mode = "recurring_template"
	ModeRecurringInstance Mode = "recurring_instance"
)

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceOddDays       RecurrenceType = "odd_days"
	RecurrenceEvenDays      RecurrenceType = "even_days"
)

type Recurrence struct {
	Type       RecurrenceType `json:"type"`
	EveryNDays int            `json:"every_n_days,omitempty"`
	DayOfMonth int            `json:"day_of_month,omitempty"`
	StartDate  *time.Time     `json:"start_date,omitempty"`
	Dates      []time.Time    `json:"dates,omitempty"`
}

type Task struct {
	ID           int64       `json:"id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	Status       Status      `json:"status"`
	Mode         Mode        `json:"mode"`
	ScheduledFor *time.Time  `json:"scheduled_for,omitempty"`
	SourceTaskID *int64      `json:"source_task_id,omitempty"`
	Recurrence   *Recurrence `json:"recurrence,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (m Mode) Valid() bool {
	switch m {
	case ModeOneTime, ModeRecurringTemplate, ModeRecurringInstance:
		return true
	default:
		return false
	}
}

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceOddDays, RecurrenceEvenDays:
		return true
	default:
		return false
	}
}
