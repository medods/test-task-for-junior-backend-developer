package recurrence

import (
	"sort"
	"time"

	"example.com/taskservice/internal/shared/dateutil"
)

type ScheduleType string

type DayParity string

const (
	ScheduleTypeEveryNDays    ScheduleType = "every_n_days"
	ScheduleTypeMonthlyDay    ScheduleType = "monthly_day"
	ScheduleTypeSpecificDates ScheduleType = "specific_dates"
	ScheduleTypeMonthParity   ScheduleType = "month_day_parity"
)

const (
	DayParityOdd  DayParity = "odd"
	DayParityEven DayParity = "even"
)

type Rule struct {
	ID            int64
	Title         string
	Description   string
	Active        bool
	ScheduleType  ScheduleType
	StartsOn      time.Time
	EndsOn        *time.Time
	EveryNDays    *int
	DayOfMonth    *int
	DayParity     *DayParity
	SpecificDates []time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (t ScheduleType) Valid() bool {
	switch t {
	case ScheduleTypeEveryNDays, ScheduleTypeMonthlyDay, ScheduleTypeSpecificDates, ScheduleTypeMonthParity:
		return true
	default:
		return false
	}
}

func (p DayParity) Valid() bool {
	switch p {
	case DayParityOdd, DayParityEven:
		return true
	default:
		return false
	}
}

func (r Rule) OccurrencesBetween(from, to time.Time) []time.Time {
	from = dateutil.Normalize(from)
	to = dateutil.Normalize(to)
	if to.Before(from) || !r.Active {
		return nil
	}

	from = dateutil.Max(from, r.StartsOn)
	if r.EndsOn != nil && r.EndsOn.Before(from) {
		return nil
	}
	if r.EndsOn != nil && r.EndsOn.Before(to) {
		to = dateutil.Normalize(*r.EndsOn)
	}
	if to.Before(from) {
		return nil
	}

	switch r.ScheduleType {
	case ScheduleTypeEveryNDays:
		return r.everyNDaysOccurrences(from, to)
	case ScheduleTypeMonthlyDay:
		return r.monthlyOccurrences(from, to)
	case ScheduleTypeSpecificDates:
		return r.specificDateOccurrences(from, to)
	case ScheduleTypeMonthParity:
		return r.monthParityOccurrences(from, to)
	default:
		return nil
	}
}

func (r Rule) everyNDaysOccurrences(from, to time.Time) []time.Time {
	if r.EveryNDays == nil || *r.EveryNDays <= 0 {
		return nil
	}

	interval := *r.EveryNDays
	daysOffset := dateutil.DaysBetween(r.StartsOn, from)
	remainder := daysOffset % interval
	if remainder < 0 {
		remainder += interval
	}

	current := from
	if remainder != 0 {
		current = dateutil.AddDays(from, interval-remainder)
	}

	occurrences := make([]time.Time, 0)
	for !current.After(to) {
		occurrences = append(occurrences, current)
		current = dateutil.AddDays(current, interval)
	}

	return occurrences
}

func (r Rule) monthlyOccurrences(from, to time.Time) []time.Time {
	if r.DayOfMonth == nil || *r.DayOfMonth < 1 || *r.DayOfMonth > 30 {
		return nil
	}

	day := *r.DayOfMonth
	cursor := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
	limit := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)
	occurrences := make([]time.Time, 0)

	for !cursor.After(limit) {
		candidate := time.Date(cursor.Year(), cursor.Month(), day, 0, 0, 0, 0, time.UTC)
		if !candidate.Before(from) && !candidate.After(to) && !candidate.Before(r.StartsOn) {
			occurrences = append(occurrences, candidate)
		}

		cursor = cursor.AddDate(0, 1, 0)
	}

	return occurrences
}

func (r Rule) specificDateOccurrences(from, to time.Time) []time.Time {
	occurrences := make([]time.Time, 0, len(r.SpecificDates))
	for _, date := range r.SpecificDates {
		normalized := dateutil.Normalize(date)
		if normalized.Before(from) || normalized.After(to) {
			continue
		}

		occurrences = append(occurrences, normalized)
	}

	sort.Slice(occurrences, func(i, j int) bool {
		return occurrences[i].Before(occurrences[j])
	})

	return occurrences
}

func (r Rule) monthParityOccurrences(from, to time.Time) []time.Time {
	if r.DayParity == nil || !r.DayParity.Valid() {
		return nil
	}

	occurrences := make([]time.Time, 0)
	for current := from; !current.After(to); current = dateutil.AddDays(current, 1) {
		day := current.Day()
		isEven := day%2 == 0
		if (*r.DayParity == DayParityEven && isEven) || (*r.DayParity == DayParityOdd && !isEven) {
			occurrences = append(occurrences, current)
		}
	}

	return occurrences
}
