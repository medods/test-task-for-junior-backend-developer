package task

import "time"

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceOddDays       RecurrenceType = "odd_days"
	RecurrenceEvenDays      RecurrenceType = "even_days"
)

type Recurrence struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      Status `json:"status"`

	Type      RecurrenceType `json:"type"`
	StartDate time.Time      `json:"start_date"`
	EndDate   *time.Time     `json:"end_date"`

	EveryNDays *int `json:"every_n_days,omitempty"`
	DayOfMonth *int `json:"day_of_month,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
