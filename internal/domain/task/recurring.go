package task

import (
	"errors"
	"fmt"
	"time"
)

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
	Type             RecurrenceType   `json:"type"`
	Interval         *int             `json:"interval,omitempty"`
	DayOfMonth       *int             `json:"day_of_month,omitempty"`
	SpecificDates    []time.Time      `json:"specific_dates,omitempty"`
	ParityMode       *ParityMode      `json:"parity_mode,omitempty"`
	StartDate        time.Time        `json:"start_date"`
	EndDate          *time.Time       `json:"end_date,omitempty"`
	Timezone         string           `json:"timezone,omitempty"`
	OverflowStrategy OverflowStrategy `json:"overflow_strategy,omitempty"`
}

var (
	ErrEmptyType               = errors.New("recurrence type cannot be empty")
	ErrZeroStartDate           = errors.New("start date cannot be zero")
	ErrEndDateBeforeStart      = errors.New("end date must be after or equal to start date")
	ErrInvalidTimezone         = errors.New("invalid timezone")
	ErrInvalidInterval         = errors.New("daily interval must be >= 1")
	ErrInvalidDayOfMonth       = errors.New("day of month must be between 1 and 31")
	ErrInvalidOverflowStrategy = errors.New("invalid overflow strategy")
	ErrEmptyDatesList          = errors.New("specific dates list cannot be empty")
	ErrZeroDateInList          = errors.New("specific dates list contains zero time")
	ErrDuplicateDate           = errors.New("specific dates list contains duplicates")
	ErrInvalidParityMode       = errors.New("parity mode must be 'even' or 'odd'")
	ErrUnknownType             = errors.New("unknown recurrence type")
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
		return ErrEmptyType
	}
	if c.StartDate.IsZero() {
		return ErrZeroStartDate
	}
	if c.EndDate != nil && c.EndDate.Before(c.StartDate) {
		return ErrEndDateBeforeStart
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
		default:
			return ErrInvalidOverflowStrategy
		}
	}
	return nil
}

func (c RecurrenceConfig) validateSpecificDates() error {
	if len(c.SpecificDates) == 0 {
		return ErrEmptyDatesList
	}
	seen := make(map[string]struct{}, len(c.SpecificDates))
	for _, d := range c.SpecificDates {
		if d.IsZero() {
			return ErrZeroDateInList
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

func (c RecurrenceConfig) CalculateEffectiveStart(now time.Time) time.Time {
	if !c.StartDate.Before(now) {
		return c.StartDate
	}

	loc := c.resolveLocation()

	switch c.Type {
	case RecurrenceTypeDaily:
		interval := 1
		if c.Interval != nil {
			interval = *c.Interval
		}
		startDay := c.StartDate.In(loc).Truncate(24 * time.Hour)
		nowDay := now.In(loc).Truncate(24 * time.Hour)
		
		daysPassed := int(nowDay.Sub(startDay)/(24*time.Hour)) + 1
		steps := (daysPassed + interval - 1) / interval 
		return startDay.AddDate(0, 0, steps*interval)

	case RecurrenceTypeMonthly:
		day := *c.DayOfMonth
		y, m, _ := now.In(loc).Date()
		check := time.Date(y, m, 1, c.StartDate.Hour(), c.StartDate.Minute(), 0, 0, loc)

		for {
			candidate := time.Date(check.Year(), check.Month(), day, check.Hour(), check.Minute(), 0, 0, loc)
			if candidate.Month() == check.Month() && !candidate.Before(now) {
				return candidate
			}
			check = check.AddDate(0, 1, 0)
		}

	case RecurrenceTypeSpecificDates:
		for _, d := range c.SpecificDates {
			if !d.Before(now) {
				return d.In(loc)
			}
		}
		return now.In(loc).Truncate(24 * time.Hour).AddDate(0, 0, 1)

	case RecurrenceTypeParity:
		mode := *c.ParityMode
		current := now.In(loc).Truncate(24 * time.Hour)
		if current.Before(now) {
			current = current.Add(24 * time.Hour)
		}
		for {
			isEven := current.Day()%2 == 0
			if (mode == ParityEven && isEven) || (mode == ParityOdd && !isEven) {
				return current
			}
			current = current.Add(24 * time.Hour)
		}

	default:
		return now.Truncate(24 * time.Hour)
	}
}

func (c RecurrenceConfig) resolveLocation() *time.Location {
	if c.Timezone != "" {
		if loc, err := time.LoadLocation(c.Timezone); err == nil {
			return loc
		}
	}
	if loc := c.StartDate.Location(); loc != nil {
		return loc
	}
	return time.UTC
}