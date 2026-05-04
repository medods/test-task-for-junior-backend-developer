package task

import (
	"errors"
	"time"
)

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
	ErrUnknownType             = errors.New("recurrence.type: unknown value")
	ErrInvalidTimezone         = errors.New("recurrence.timezone: invalid IANA format")
)

type RequiredTypeSpec struct{}

func (s RequiredTypeSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Type == "" {
		return ErrTypeRequired
	}
	return nil
}

type StartDateRequiredSpec struct{}

func (s StartDateRequiredSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.StartDate.IsZero() {
		return ErrStartDateRequired
	}
	return nil
}

type EndDateValidSpec struct{}

func (s EndDateValidSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.EndDate != nil && cfg.EndDate.Before(cfg.StartDate) {
		return ErrStartDateAfterEnd
	}
	return nil
}

type TimezoneValidSpec struct{}

func (s TimezoneValidSpec) IsSatisfiedBy(cfg RecurrenceConfig) error {
	if cfg.Timezone != "" {
		if _, err := time.LoadLocation(cfg.Timezone); err != nil {
			return ErrInvalidTimezone
		}
	}
	return nil
}