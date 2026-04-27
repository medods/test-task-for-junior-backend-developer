package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RepeatInterval string

const (
	RepeatNever   RepeatInterval = "never"
	RepeatDaily   RepeatInterval = "daily"
	RepeatMonthly RepeatInterval = "monthly"
	RepeatCustom  RepeatInterval = "custom"
	RepeatEven    RepeatInterval = "even_days"
	RepeatOdd     RepeatInterval = "odd_days"
)

type RepeatConfig struct {
	DailyInterval *int        `json:"daily_interval,omitempty"`
	MonthDay      *int        `json:"month_day,omitempty"`
	SpecificDates []time.Time `json:"specific_dates,omitempty"`
}

type Task struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Repeat      RepeatInterval `json:"repeat"`
	Config      *RepeatConfig  `json:"config,omitempty"`
	Status      Status         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (r RepeatInterval) Valid() bool {
	switch r {
	case RepeatNever, RepeatDaily, RepeatMonthly, RepeatCustom, RepeatEven, RepeatOdd:
		return true
	default:
		return false
	}
}
