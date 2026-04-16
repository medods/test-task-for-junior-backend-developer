package task

import (
	"crypto/rand"
	"fmt"
	"sort"
	"time"
)

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenOdd       RecurrenceType = "even_odd"
)

const maxInstances = 366

type RecurrenceConfig struct {
	Type        RecurrenceType `json:"type"`
	EveryNDays  int            `json:"every_n_days,omitempty"`
	DaysOfMonth []int          `json:"days_of_month,omitempty"`
	Parity      string         `json:"parity,omitempty"`
}

type RecurrenceInput struct {
	Type        RecurrenceType
	EveryNDays  int
	DaysOfMonth []int
	Parity      string
	Dates       []time.Time
	StartDate   time.Time
	EndDate     time.Time
}

func (r RecurrenceInput) Config() *RecurrenceConfig {
	return &RecurrenceConfig{
		Type:        r.Type,
		EveryNDays:  r.EveryNDays,
		DaysOfMonth: r.DaysOfMonth,
		Parity:      r.Parity,
	}
}

func ExpandDates(input RecurrenceInput) ([]time.Time, error) {
	switch input.Type {
	case RecurrenceDaily:
		return expandDaily(input)
	case RecurrenceMonthly:
		return expandMonthly(input)
	case RecurrenceSpecificDates:
		return expandSpecificDates(input)
	case RecurrenceEvenOdd:
		return expandEvenOdd(input)
	default:
		return nil, fmt.Errorf("unknown recurrence type %q; valid values: daily, monthly, specific_dates, even_odd", input.Type)
	}
}

func NewGroupID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func midnight(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func expandDaily(s RecurrenceInput) ([]time.Time, error) {
	if s.EveryNDays <= 0 {
		return nil, fmt.Errorf("every_n_days must be >= 1")
	}
	if err := requireDateRange(s); err != nil {
		return nil, err
	}

	start := midnight(s.StartDate)
	end := midnight(s.EndDate)

	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, s.EveryNDays) {
		dates = append(dates, d)
		if len(dates) > maxInstances {
			return nil, tooManyInstancesErr()
		}
	}
	return dates, nil
}

func expandMonthly(s RecurrenceInput) ([]time.Time, error) {
	if len(s.DaysOfMonth) == 0 {
		return nil, fmt.Errorf("days_of_month is required for monthly recurrence")
	}
	for _, d := range s.DaysOfMonth {
		if d < 1 || d > 30 {
			return nil, fmt.Errorf("days_of_month values must be between 1 and 30, got %d", d)
		}
	}
	if err := requireDateRange(s); err != nil {
		return nil, err
	}

	start := midnight(s.StartDate)
	end := midnight(s.EndDate)

	curMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	var dates []time.Time
	for !curMonth.After(endMonth) {
		for _, day := range s.DaysOfMonth {
			candidate := time.Date(curMonth.Year(), curMonth.Month(), day, 0, 0, 0, 0, time.UTC)
			if candidate.Month() != curMonth.Month() {
				continue
			}
			if candidate.Before(start) || candidate.After(end) {
				continue
			}
			dates = append(dates, candidate)
			if len(dates) > maxInstances {
				return nil, tooManyInstancesErr()
			}
		}
		curMonth = curMonth.AddDate(0, 1, 0)
	}

	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates, nil
}

func expandSpecificDates(s RecurrenceInput) ([]time.Time, error) {
	if len(s.Dates) == 0 {
		return nil, fmt.Errorf("dates is required for specific_dates recurrence")
	}
	if len(s.Dates) > maxInstances {
		return nil, tooManyInstancesErr()
	}

	seen := make(map[string]bool, len(s.Dates))
	var dates []time.Time
	for _, d := range s.Dates {
		norm := midnight(d)
		key := norm.Format("2006-01-02")
		if !seen[key] {
			seen[key] = true
			dates = append(dates, norm)
		}
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates, nil
}

func expandEvenOdd(s RecurrenceInput) ([]time.Time, error) {
	if s.Parity != "even" && s.Parity != "odd" {
		return nil, fmt.Errorf("parity must be \"even\" or \"odd\", got %q", s.Parity)
	}
	if err := requireDateRange(s); err != nil {
		return nil, err
	}

	start := midnight(s.StartDate)
	end := midnight(s.EndDate)

	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		isEven := d.Day()%2 == 0
		if (s.Parity == "even" && isEven) || (s.Parity == "odd" && !isEven) {
			dates = append(dates, d)
			if len(dates) > maxInstances {
				return nil, tooManyInstancesErr()
			}
		}
	}
	return dates, nil
}

func requireDateRange(s RecurrenceInput) error {
	if s.StartDate.IsZero() || s.EndDate.IsZero() {
		return fmt.Errorf("start_date and end_date are required for %s recurrence", s.Type)
	}
	if midnight(s.EndDate).Before(midnight(s.StartDate)) {
		return fmt.Errorf("end_date must be >= start_date")
	}
	return nil
}

func tooManyInstancesErr() error {
	return fmt.Errorf("recurrence would generate more than %d instances; narrow the date range or increase the interval", maxInstances)
}
