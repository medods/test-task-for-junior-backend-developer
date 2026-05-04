package task

import (
	"fmt"
	"time"
)

func (c RecurrenceConfig) Validate() error {
	if err := c.validateCommon(); err != nil {
		return err
	}

	switch c.Type {
	case RecurrenceTypeDaily:
		return c.validateDaily()
	case RecurrenceTypeMonthly:
		return c.validateMonthly()
	case RecurrenceTypeSpecificDates:
		return c.validateSpecificDates()
	case RecurrenceTypeParity:
		return c.validateParity()
	default:
		return ErrUnknownType
	}
}

func (c RecurrenceConfig) validateCommon() error {
	if c.Type == "" {
		return ErrTypeRequired
	}

	if c.StartDate.IsZero() {
		return ErrStartDateRequired
	}

	if c.EndDate != nil {
		if c.EndDate.Before(c.StartDate) {
			return ErrStartDateAfterEnd
		}
	}

	if c.Timezone != "" {
		if _, err := time.LoadLocation(c.Timezone); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidTimezone, err)
		}
	}

	return nil
}

func (c RecurrenceConfig) validateDaily() error {
	if c.Interval == nil || *c.Interval < 1 {
		return ErrInvalidInterval
	}
	return nil
}

func (c RecurrenceConfig) validateMonthly() error {
	if c.DayOfMonth == nil || *c.DayOfMonth < 1 || *c.DayOfMonth > 31 {
		return ErrInvalidDayOfMonth
	}

	if c.OverflowStrategy != "" {
		switch c.OverflowStrategy {
		case OverflowSkip, OverflowLastDay, OverflowNextMonth:
			// валидно
		default:
			return ErrInvalidOverflowStrategy
		}
	}

	return nil
}

func (c RecurrenceConfig) validateSpecificDates() error {
	if len(c.SpecificDates) == 0 {
		return ErrEmptyDates
	}

	seen := make(map[string]struct{}, len(c.SpecificDates))
	for _, d := range c.SpecificDates {
		if d.IsZero() {
			return ErrInvalidDateInList
		}

		key := d.UTC().Format("2006-01-02")
		if _, exists := seen[key]; exists {
			return ErrDuplicateDate
		}
		seen[key] = struct{}{}
	}

	return nil
}

func (c RecurrenceConfig) validateParity() error {
	if c.ParityMode == nil {
		return ErrInvalidParityMode
	}

	if *c.ParityMode != ParityEven && *c.ParityMode != ParityOdd {
		return ErrInvalidParityMode
	}

	return nil
}