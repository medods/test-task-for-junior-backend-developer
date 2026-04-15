// Package recurrence
package recurrence

type RecurrenceType string

const (
	NoneRecurrence RecurrenceType = "none"
	DailyEveryN    RecurrenceType = "daily_every_n"
	MonthlyDay     RecurrenceType = "monthly_day"
	SpecificDates  RecurrenceType = "specific_dates"
	MonthParity    RecurrenceType = "month_parity"
)
