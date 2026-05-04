package task

import "time"

type RecurrenceType string

const (
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeParity        RecurrenceType = "parity"
)

type ParityMode string

const (
	ParityEven ParityMode = "even"
	ParityOdd  ParityMode = "odd"
)

type OverflowStrategy string

const (
	OverflowSkip      OverflowStrategy = "skip"     
	OverflowLastDay   OverflowStrategy = "last_day"  
	OverflowNextMonth OverflowStrategy = "next_month" 
)

type RecurrenceConfig struct {
	Type RecurrenceType `json:"type"`
	Interval *int `json:"interval,omitempty"`
	DayOfMonth *int `json:"day_of_month,omitempty"`
	SpecificDates []time.Time `json:"specific_dates,omitempty"`
	ParityMode *ParityMode `json:"parity_mode,omitempty"`
	StartDate time.Time `json:"start_date"`
	EndDate *time.Time `json:"end_date,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	OverflowStrategy OverflowStrategy `json:"overflow_strategy,omitempty"`
}