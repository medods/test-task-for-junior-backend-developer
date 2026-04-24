package task

import (
	"errors"
	"slices"
	"time"
)

var ErrInvalidPeriod = errors.New("invalid task period")

type PeriodType string

const (
	PeriodDaily    PeriodType = "daily"
	PeriodMonthly  PeriodType = "monthly"
	PeriodSpecific PeriodType = "specific"
	PeriodEvenOdd  PeriodType = "even_odd"
)

type EvenOddType string

const (
	Even EvenOddType = "even"
	Odd  EvenOddType = "odd"
)

type Period struct {
	Type          PeriodType   `json:"type"`
	IntervalDays  *int         `json:"interval_days,omitempty"`
	MonthDay      *int         `json:"month_day,omitempty"`
	SpecificDates []time.Time  `json:"specific_dates,omitempty"`
	EvenOddType   *EvenOddType `json:"even_odd_type,omitempty"`
}

func (p *Period) Clone() *Period {
	if p == nil {
		return nil
	}

	clone := *p
	if len(p.SpecificDates) > 0 {
		clone.SpecificDates = append([]time.Time(nil), p.SpecificDates...)
	}

	if p.IntervalDays != nil {
		value := *p.IntervalDays
		clone.IntervalDays = &value
	}

	if p.MonthDay != nil {
		value := *p.MonthDay
		clone.MonthDay = &value
	}

	if p.EvenOddType != nil {
		value := *p.EvenOddType
		clone.EvenOddType = &value
	}

	return &clone
}

func (p *Period) Normalize() (*Period, error) {
	if p == nil {
		return nil, nil
	}

	normalized := p.Clone()
	switch normalized.Type {
	case PeriodDaily:
		if normalized.IntervalDays == nil || *normalized.IntervalDays < 1 {
			return nil, ErrInvalidPeriod
		}
		normalized.MonthDay = nil
		normalized.SpecificDates = nil
		normalized.EvenOddType = nil
	case PeriodMonthly:
		if normalized.MonthDay == nil || *normalized.MonthDay < 1 || *normalized.MonthDay > 30 {
			return nil, ErrInvalidPeriod
		}
		normalized.IntervalDays = nil
		normalized.SpecificDates = nil
		normalized.EvenOddType = nil
	case PeriodSpecific:
		if len(normalized.SpecificDates) == 0 {
			return nil, ErrInvalidPeriod
		}

		dates := make([]time.Time, 0, len(normalized.SpecificDates))
		for _, value := range normalized.SpecificDates {
			dates = append(dates, normalizeDate(value))
		}
		slices.SortFunc(dates, func(a, b time.Time) int {
			return a.Compare(b)
		})
		dates = slices.CompactFunc(dates, func(a, b time.Time) bool {
			return a.Equal(b)
		})

		normalized.IntervalDays = nil
		normalized.MonthDay = nil
		normalized.SpecificDates = dates
		normalized.EvenOddType = nil
	case PeriodEvenOdd:
		if normalized.EvenOddType == nil || !normalized.EvenOddType.Valid() {
			return nil, ErrInvalidPeriod
		}
		normalized.IntervalDays = nil
		normalized.MonthDay = nil
		normalized.SpecificDates = nil
	default:
		return nil, ErrInvalidPeriod
	}

	return normalized, nil
}

func (p *Period) NextScheduledDate(now time.Time, anchor time.Time) *time.Time {
	if p == nil {
		return nil
	}

	current := normalizeDate(now)
	anchorDate := normalizeDate(anchor)

	switch p.Type {
	case PeriodDaily:
		return nextDailyDate(current, anchorDate, *p.IntervalDays)
	case PeriodMonthly:
		return nextMonthlyDate(current, *p.MonthDay)
	case PeriodSpecific:
		for _, value := range p.SpecificDates {
			if !value.Before(current) {
				date := value
				return &date
			}
		}
		return nil
	case PeriodEvenOdd:
		return nextEvenOddDate(current, *p.EvenOddType)
	default:
		return nil
	}
}

func (e EvenOddType) Valid() bool {
	return e == Even || e == Odd
}

func normalizeDate(value time.Time) time.Time {
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func nextDailyDate(current time.Time, anchor time.Time, interval int) *time.Time {
	if !current.After(anchor) {
		date := anchor
		return &date
	}

	diffDays := int(current.Sub(anchor).Hours() / 24)
	remainder := diffDays % interval
	if remainder == 0 {
		date := current
		return &date
	}

	date := current.AddDate(0, 0, interval-remainder)
	return &date
}

func nextMonthlyDate(current time.Time, day int) *time.Time {
	year := current.Year()
	month := current.Month()

	for i := 0; i < 24; i++ {
		if day <= daysInMonth(year, month) {
			candidate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			if !candidate.Before(current) {
				return &candidate
			}
		}

		year, month = nextMonth(year, month)
	}

	return nil
}

func nextEvenOddDate(current time.Time, parity EvenOddType) *time.Time {
	for offset := 0; offset <= 31; offset++ {
		candidate := current.AddDate(0, 0, offset)
		if parity == Even && candidate.Day()%2 == 0 {
			return &candidate
		}
		if parity == Odd && candidate.Day()%2 == 1 {
			return &candidate
		}
	}

	return nil
}

func nextMonth(year int, month time.Month) (int, time.Month) {
	if month == time.December {
		return year + 1, time.January
	}

	return year, month + 1
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
