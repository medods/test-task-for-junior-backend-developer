package task

import (
	"fmt"
	"time"
)

const defaultDayStep = 1

type RecurrenceType string

const (
	Daily         RecurrenceType = "daily"
	Monthly       RecurrenceType = "monthly"
	SpecificDates RecurrenceType = "specific_dates"
	EvenOdd       RecurrenceType = "even_odd"
)

type ParityWeekType string

const (
	Even ParityWeekType = "even"
	Odd  ParityWeekType = "odd"
)

var recurrenceGenerators = map[RecurrenceType]func(RecurrenceInput) ([]time.Time, error){
	Daily:         generateDailyDates,
	Monthly:       generateMonthlyDates,
	SpecificDates: func(input RecurrenceInput) ([]time.Time, error) { return input.Dates, nil },
	EvenOdd:       generateEvenOddDates,
}

func generateDates(input RecurrenceInput) ([]time.Time, error) {
	generator, ok := recurrenceGenerators[RecurrenceType(input.Type)]
	if !ok {
		return nil, fmt.Errorf("recurrence type not valid: %s", input.Type)
	}
	return generator(input)
}

func generateDailyDates(input RecurrenceInput) ([]time.Time, error) {
	step := defaultDayStep
	if input.EveryNDays != nil && *input.EveryNDays > 0 {
		step = *input.EveryNDays
	}

	var dates []time.Time
	current := input.StartDate

	for !current.After(input.EndDate) {
		dates = append(dates, current)
		current = current.AddDate(0, 0, step)
	}
	return dates, nil
}

func generateMonthlyDates(input RecurrenceInput) ([]time.Time, error) {
	day := *input.DayOfMonth

	var dates []time.Time
	current := firstMonthDayOnOrAfter(input.StartDate, day)

	for !current.After(input.EndDate) {
		dates = append(dates, current)
		current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
		current = adjustDay(current, day)
	}

	return dates, nil
}

func generateEvenOddDates(input RecurrenceInput) ([]time.Time, error) {
	parity := *input.Parity

	var dates []time.Time
	current := input.StartDate

	for !current.After(input.EndDate) {
		day := current.Day()
		if (parity == string(Even) && day%2 == 0) || (parity == string(Odd) && day%2 != 0) {
			dates = append(dates, current)
		}

		current = current.AddDate(0, 0, 1)
	}
	return dates, nil
}

func firstMonthDayOnOrAfter(start time.Time, day int) time.Time {
	candidate := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	candidate = adjustDay(candidate, day)

	if candidate.Before(start) {
		nextMonth := start.AddDate(0, 1, 0)
		candidate = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, start.Location())
		candidate = adjustDay(candidate, day)
	}

	return candidate
}

func adjustDay(t time.Time, day int) time.Time {
	year, month, _ := t.Date()
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location()).Day()
	actualDay := min(day, lastDay)
	return time.Date(year, month, actualDay, 0, 0, 0, 0, t.Location())
}
