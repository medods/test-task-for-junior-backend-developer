package recurrence

import (
	"example.com/taskservice/internal/domain/common"
)

// RecurrencePayload contains parameters for the selected RecurrenceType.
type RecurrencePayload struct {
	IntervalDays  int               `json:"interval_days" validate:"omitempty,min=1"`
	MonthDay      int               `json:"month_day" validate:"omitempty,min=1,max=30"`
	SpecificDates []common.DateType `json:"specific_dates" validate:"omitempty,min=1"`
	Parity        Parity            `json:"parity" validate:"omitempty,oneof=none even odd"`
}
