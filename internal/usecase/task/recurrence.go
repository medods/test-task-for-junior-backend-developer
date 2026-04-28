package task

import (
	"fmt"
	"sort"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func validateAndNormalizeRecurrence(recurrence *taskdomain.Recurrence) (*taskdomain.Recurrence, error) {
	if err := validateRecurrence(recurrence); err != nil {
		return nil, err
	}

	return normalizeRecurrence(recurrence), nil
}

func validateRecurrence(recurrence *taskdomain.Recurrence) error {
	if recurrence == nil {
		return fmt.Errorf("%w: recurrence is required", ErrInvalidInput)
	}

	if !recurrence.Type.Valid() {
		return fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	switch recurrence.Type {
	case taskdomain.RecurrenceDaily:
		if recurrence.EveryNDays <= 0 {
			return fmt.Errorf("%w: every_n_days must be positive", ErrInvalidInput)
		}

		if recurrence.StartDate == nil || recurrence.StartDate.IsZero() {
			return fmt.Errorf("%w: start_date is required for daily recurrence", ErrInvalidInput)
		}

		if recurrence.DayOfMonth != 0 {
			return fmt.Errorf("%w: day_of_month is not allowed for daily recurrence", ErrInvalidInput)
		}

		if len(recurrence.Dates) > 0 {
			return fmt.Errorf("%w: dates are not allowed for daily recurrence", ErrInvalidInput)
		}

	case taskdomain.RecurrenceMonthly:
		if recurrence.EveryNDays != 0 {
			return fmt.Errorf("%w: every_n_days is not allowed for monthly recurrence", ErrInvalidInput)
		}

		if recurrence.DayOfMonth < 1 || recurrence.DayOfMonth > 30 {
			return fmt.Errorf("%w: day_of_month must be between 1 and 30", ErrInvalidInput)
		}

		if recurrence.StartDate == nil || recurrence.StartDate.IsZero() {
			return fmt.Errorf("%w: start_date is required for monthly recurrence", ErrInvalidInput)
		}

		if len(recurrence.Dates) > 0 {
			return fmt.Errorf("%w: dates are not allowed for monthly recurrence", ErrInvalidInput)
		}

	case taskdomain.RecurrenceSpecificDates:
		if recurrence.EveryNDays != 0 {
			return fmt.Errorf("%w: every_n_days is not allowed for specific dates recurrence", ErrInvalidInput)
		}

		if recurrence.DayOfMonth != 0 {
			return fmt.Errorf("%w: day_of_month is not allowed for specific dates recurrence", ErrInvalidInput)
		}

		if recurrence.StartDate != nil {
			return fmt.Errorf("%w: start_date is not allowed for specific dates recurrence", ErrInvalidInput)
		}

		if len(recurrence.Dates) == 0 {
			return fmt.Errorf("%w: dates are required for specific dates recurrence", ErrInvalidInput)
		}

		if err := validateUniqueDates(recurrence.Dates); err != nil {
			return err
		}

	case taskdomain.RecurrenceOddDays:
		if recurrence.EveryNDays != 0 {
			return fmt.Errorf("%w: every_n_days is not allowed for odd days recurrence", ErrInvalidInput)
		}

		if recurrence.DayOfMonth != 0 {
			return fmt.Errorf("%w: day_of_month is not allowed for odd days recurrence", ErrInvalidInput)
		}

		if recurrence.StartDate == nil || recurrence.StartDate.IsZero() {
			return fmt.Errorf("%w: start_date is required for odd days recurrence", ErrInvalidInput)
		}

		if len(recurrence.Dates) > 0 {
			return fmt.Errorf("%w: dates are not allowed for odd days recurrence", ErrInvalidInput)
		}

	case taskdomain.RecurrenceEvenDays:
		if recurrence.EveryNDays != 0 {
			return fmt.Errorf("%w: every_n_days is not allowed for even days recurrence", ErrInvalidInput)
		}

		if recurrence.DayOfMonth != 0 {
			return fmt.Errorf("%w: day_of_month is not allowed for even days recurrence", ErrInvalidInput)
		}

		if recurrence.StartDate == nil || recurrence.StartDate.IsZero() {
			return fmt.Errorf("%w: start_date is required for even days recurrence", ErrInvalidInput)
		}

		if len(recurrence.Dates) > 0 {
			return fmt.Errorf("%w: dates are not allowed for even days recurrence", ErrInvalidInput)
		}
	}

	return nil
}

func validateUniqueDates(dates []time.Time) error {
	seen := make(map[string]struct{}, len(dates))

	for _, value := range dates {
		if value.IsZero() {
			return fmt.Errorf("%w: dates must not contain zero value", ErrInvalidInput)
		}

		key := dateOnly(value).Format(time.DateOnly)

		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: dates must be unique", ErrInvalidInput)
		}

		seen[key] = struct{}{}
	}

	return nil
}

func normalizeRecurrence(recurrence *taskdomain.Recurrence) *taskdomain.Recurrence {
	if recurrence == nil {
		return nil
	}

	normalized := *recurrence

	if recurrence.StartDate != nil {
		startDate := dateOnly(*recurrence.StartDate)
		normalized.StartDate = &startDate
	}

	if len(recurrence.Dates) > 0 {
		normalized.Dates = make([]time.Time, 0, len(recurrence.Dates))

		for _, value := range recurrence.Dates {
			normalized.Dates = append(normalized.Dates, dateOnly(value))
		}

		sort.Slice(normalized.Dates, func(i, j int) bool {
			return normalized.Dates[i].Before(normalized.Dates[j])
		})
	}

	return &normalized
}

func generateScheduledDates(
	recurrence *taskdomain.Recurrence,
	from time.Time,
	horizonDays int,
) []time.Time {
	if recurrence == nil || horizonDays <= 0 {
		return nil
	}

	recurrence = normalizeRecurrence(recurrence)

	start := dateOnly(from)
	endExclusive := start.AddDate(0, 0, horizonDays)

	switch recurrence.Type {
	case taskdomain.RecurrenceDaily:
		return generateDailyDates(recurrence, start, endExclusive)

	case taskdomain.RecurrenceMonthly:
		return generateMonthlyDates(recurrence, start, endExclusive)

	case taskdomain.RecurrenceSpecificDates:
		return generateSpecificDates(recurrence, start, endExclusive)

	case taskdomain.RecurrenceOddDays:
		return generateOddEvenDates(recurrence, start, endExclusive, true)

	case taskdomain.RecurrenceEvenDays:
		return generateOddEvenDates(recurrence, start, endExclusive, false)

	default:
		return nil
	}
}

func generateDailyDates(
	recurrence *taskdomain.Recurrence,
	from time.Time,
	endExclusive time.Time,
) []time.Time {
	if recurrence.StartDate == nil || recurrence.EveryNDays <= 0 {
		return nil
	}

	recurrenceStart := dateOnly(*recurrence.StartDate)

	var dates []time.Time

	for current := from; current.Before(endExclusive); current = current.AddDate(0, 0, 1) {
		if current.Before(recurrenceStart) {
			continue
		}

		daysFromStart := int(current.Sub(recurrenceStart).Hours() / 24)
		if daysFromStart%recurrence.EveryNDays == 0 {
			dates = append(dates, current)
		}
	}

	return dates
}

func generateMonthlyDates(
	recurrence *taskdomain.Recurrence,
	from time.Time,
	endExclusive time.Time,
) []time.Time {
	if recurrence.DayOfMonth < 1 || recurrence.DayOfMonth > 30 {
		return nil
	}

	if recurrence.StartDate == nil {
		return nil
	}

	recurrenceStart := dateOnly(*recurrence.StartDate)
	currentMonth := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)

	var dates []time.Time

	for currentMonth.Before(endExclusive) {
		candidate := clampedDateInMonth(
			currentMonth.Year(),
			currentMonth.Month(),
			recurrence.DayOfMonth,
		)

		if !candidate.Before(from) &&
			candidate.Before(endExclusive) &&
			!candidate.Before(recurrenceStart) {
			dates = append(dates, candidate)
		}

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return dates
}

func generateSpecificDates(
	recurrence *taskdomain.Recurrence,
	from time.Time,
	endExclusive time.Time,
) []time.Time {
	dates := make([]time.Time, 0, len(recurrence.Dates))

	for _, value := range recurrence.Dates {
		date := dateOnly(value)

		if date.Before(from) || !date.Before(endExclusive) {
			continue
		}

		dates = append(dates, date)
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return dates
}

func generateOddEvenDates(
	recurrence *taskdomain.Recurrence,
	from time.Time,
	endExclusive time.Time,
	odd bool,
) []time.Time {
	if recurrence.StartDate == nil {
		return nil
	}

	recurrenceStart := dateOnly(*recurrence.StartDate)

	var dates []time.Time

	for current := from; current.Before(endExclusive); current = current.AddDate(0, 0, 1) {
		if current.Before(recurrenceStart) {
			continue
		}

		isOddDay := current.Day()%2 != 0
		if isOddDay == odd {
			dates = append(dates, current)
		}
	}

	return dates
}
