package recurrence

import (
	"example.com/taskservice/internal/domain/common"
)

type RecurrencePayload struct {
	IntervalDays  int               `json:"interval_days" validate:"min=1"`
	MonthDay      int               `json:"month_day" validate:"min=1,max=30"`
	SpecificDates []common.DateType `json:"specific_dates"`
	Parity        Parity            `json:"parity" validate:"oneof=none even odd"`
}
