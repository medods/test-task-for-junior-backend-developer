package task

import "errors"

var (
	ErrTypeRequired            = errors.New("recurrence.type: required")
	ErrStartDateRequired       = errors.New("recurrence.start_date: required")
	ErrStartDateAfterEnd       = errors.New("recurrence.dates: start_date must be <= end_date")
	ErrInvalidInterval         = errors.New("recurrence.interval: must be >= 1")
	ErrInvalidDayOfMonth       = errors.New("recurrence.day_of_month: must be between 1 and 31")
	ErrInvalidOverflowStrategy = errors.New("recurrence.overflow_strategy: invalid value")
	ErrEmptyDates              = errors.New("recurrence.specific_dates: cannot be empty")
	ErrDuplicateDate           = errors.New("recurrence.specific_dates: contains duplicates")
	ErrInvalidDateInList       = errors.New("recurrence.specific_dates: contains zero or invalid dates")
	ErrInvalidParityMode       = errors.New("recurrence.parity_mode: must be 'even' or 'odd'")
	ErrInvalidTimezone         = errors.New("recurrence.timezone: invalid IANA format")
)

var (
	ErrUnknownType = errors.New("recurrence.type: unknown value")
)

func GetErrorCategory(err error) string {
	switch {
	case errors.Is(err, ErrTypeRequired),
		errors.Is(err, ErrStartDateRequired),
		errors.Is(err, ErrStartDateAfterEnd),
		errors.Is(err, ErrInvalidInterval),
		errors.Is(err, ErrInvalidDayOfMonth),
		errors.Is(err, ErrInvalidOverflowStrategy),
		errors.Is(err, ErrEmptyDates),
		errors.Is(err, ErrDuplicateDate),
		errors.Is(err, ErrInvalidDateInList),
		errors.Is(err, ErrInvalidParityMode),
		errors.Is(err, ErrInvalidTimezone):
		return "validation"

	case errors.Is(err, ErrUnknownType):
		return "logic"

	default:
		return "unknown"
	}
}