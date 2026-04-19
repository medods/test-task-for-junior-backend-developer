package task

import (
	"fmt"
	"time"
)

const (
	defaultDayStep = 1

	daily         = "daily"
	monthly       = "monthly"
	specificDates = "specific_dates"
	evenOdd       = "even_odd"
)

func generateDates(input RecurrenceInput) ([]time.Time, error) {
	switch input.Type {
	case daily:
		return generateDailyDates(input)
	case monthly:
		return generateMonthlyDates(input)
	case specificDates:
		return input.Dates, nil
	case evenOdd:
		return generateEvenOddDates(input)
	default:
		return nil, fmt.Errorf("неизвестный тип периодичности: %s", input.Type)
	}
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
	start := input.StartDate
	end := input.EndDate

	var dates []time.Time

	current := firstMonthDayOnOrAfter(start, day)
	for !current.After(end) {
		dates = append(dates, current)
		current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
		current = adjustDay(current, day)
	}

	return dates, nil
}

func generateEvenOddDates(input RecurrenceInput) ([]time.Time, error) {
	parity := *input.Parity
	start := input.StartDate
	end := input.EndDate

	var dates []time.Time
	current := start

	for !current.After(end) {
		day := current.Day()
		if (parity == "even" && day%2 == 0) || (parity == "odd" && day%2 != 0) {
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
