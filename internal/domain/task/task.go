package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string

const (
	RecurrenceTypeNone          RecurrenceType = "none"
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeMonthParity   RecurrenceType = "month_day_parity"
)

type MonthDayParity string

const (
	MonthDayParityEven MonthDayParity = "even"
	MonthDayParityOdd  MonthDayParity = "odd"
)

type Recurrence struct {
	Type       RecurrenceType  `json:"type"`
	EveryNDays *int            `json:"every_n_days,omitempty"`
	DayOfMonth *int            `json:"day_of_month,omitempty"`
	Dates      []string        `json:"dates,omitempty"`
	Parity     *MonthDayParity `json:"parity,omitempty"`
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

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (t RecurrenceType) Valid() bool {
	switch t {
	case RecurrenceTypeNone, RecurrenceTypeDaily, RecurrenceTypeMonthly, RecurrenceTypeSpecificDates, RecurrenceTypeMonthParity:
		return true
	default:
		return false
	}
}

func (p MonthDayParity) Valid() bool {
	switch p {
	case MonthDayParityEven, MonthDayParityOdd:
		return true
	default:
		return false
	}
}
